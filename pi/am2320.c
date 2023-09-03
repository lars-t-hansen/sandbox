/* -*- fill-column: 100 -*-

   Linux implementation of the Interface to the Adafruit AM2320 family of temperature/humidity
   sensors.

   The following code has been copied from https://github.com/Gozem/am2320 (MIT License).  It has
   been adapted for robustness (retry wakeup) and modularity (header file), edited for clarity, and
   expanded with new functionality.

   Possible TODO items:

   - the first read after bootup tends to yield the value 0, it's unclear what the reason
     for this is, but it would be nice to address it.

   - in principle, reads and writes may be partial, but it's unclear what this would mean
     in the context of the device, so I've not bothered to change that.

   - there may be other errors than EREMOTEIO that should cause retries

   - there may be reason to retry also the read if it fails (read error, prefix error, crc error)?
*/

#include <stdio.h>
#include <string.h>
#include <sys/ioctl.h>
#include <fcntl.h> 
#include <linux/i2c-dev.h>
#include <unistd.h>
#include <errno.h>

#include "am2320.h"

static uint16_t 
calc_crc16(const uint8_t *buf, size_t len) {
  uint16_t crc = 0xFFFF;
  
  while(len--) {
    crc ^= (uint16_t) *buf++;
    for (unsigned i = 0; i < 8; i++) {
      if (crc & 0x0001) {
	crc >>= 1;
	crc ^= 0xA001;
      } else {
	crc >>= 1;      
      }
    }
  }
  
  return crc;
}

static uint16_t
combine_bytes(uint8_t msb, uint8_t lsb)
{
  return ((uint16_t)msb << 8) | (uint16_t)lsb;
}

int
am2320_open(unsigned i2c_device_no, am2320_t* out_fd) {
  char device[64];
  sprintf(device, "/dev/i2c-%u", i2c_device_no);

  int fd = open(device, O_RDWR);
  if (fd < 0)
    return AM2320_ERR_OPEN;

  if (ioctl(fd, I2C_SLAVE, AM2320_ADDRESS) < 0) {
    close(fd);
    return AM2320_ERR_INIT;
  }
   
  *out_fd = fd;
  return AM2320_OK;
}

void
am2320_close(am2320_t fd) {
  close(fd);
}

/* buf must be large enough to hold numregs bytes */
static int
wake_and_write(int fd, uint8_t firstreg, uint8_t numregs, uint8_t* buf) {

  /* wake AM2320 up, goes to sleep to not warm up and affect the humidity sensor */
  int iter = 0;
 again:
  write(fd, NULL, 0);
  usleep(1000); /* at least 0.8ms, at most 3ms */
  
  /* Max 10 registers + prefix + crc */
  uint8_t data[15];
  data[0] = 0x10;
  data[1] = firstreg;
  data[2] = numregs;
  memcpy(data+3, buf, numregs);
  // FIXME
  // CRC - not sure what we're computing across, if it's the entire buffer or not
  uint16_t crcdata = calc_crc16(data, numregs + 2);
  data[3+numregs] = crcdata;
  data[4+numregs] = crcdata >> 8
  if (write(fd, data, numregs+4) < 0) {
    if (errno == EREMOTEIO && ++iter < 5) {
      /* Assumes that no bytes were written / that writes are idempotent, which should be OK */
      goto again;
    }
    return AM2320_ERR_WARMUP;
  }

  // TODO: There is a response which we should read and decode!

  return AM2320_OK;
}  

/* buf must be large enough to hold numregs bytes */
static int
wake_and_read(int fd, uint8_t firstreg, uint8_t numregs, uint8_t* buf) {

  /* wake AM2320 up, goes to sleep to not warm up and affect the humidity sensor */
  int iter = 0;
 again:
  write(fd, NULL, 0);
  usleep(1000); /* at least 0.8ms, at most 3ms */
  
  /* signal we want to read */
  uint8_t setup_buf[3] = {0x03, firstreg, numregs};
  if (write(fd, setup_buf, sizeof(setup_buf)) < 0) {
    if (errno == EREMOTEIO && ++iter < 5) {
      goto again;
    }
    return AM2320_ERR_WARMUP;
  }
  
  /* wait for AM2320 */
  usleep(1600); /* Wait atleast 1.5ms */

  /* Max length of returned value is 4 + 10 = 14 bytes */
  uint8_t tmp[14];
  if (read(fd, tmp, 4 + numregs) < 0)
    return AM2320_ERR_READ;

  if (tmp[0] != 0x03 || tmp[1] != numregs) {
    // TODO: Decode error message
    return AM2320_ERR_PREFIX;
  }

  /* Check CRC - in last two bytes, little-endian (weird but true) */
  uint16_t crcdata = calc_crc16(tmp, numregs + 2);
  uint16_t crcread = combine_bytes(tmp[numregs + 3], tmp[numregs + 2]);
  if (crcdata != crcread) 
    return AM2320_ERR_CRC;
  
  memcpy(buf, tmp+2, numregs);
  return AM2320_OK;
}

int
am2320_read_id(am2320_t fd, int* model, int* version, unsigned* dev_id) {
  int res;
  uint8_t data[7];

  if ((res = wake_and_read(fd, 0x08, 7, data)) != AM2320_OK)
    return res;

  *model = combine_bytes(data[0], data[1]);
  *version = data[2];
  *dev_id = ((unsigned)data[3] << 24) | ((unsigned)data[4] << 16) | ((unsigned)data[5] << 8) | (unsigned)data[6];

  return AM2320_OK;
}

int 
am2320_read_sensors(am2320_t fd, float *out_temperature, float *out_humidity) 
{
  int res;

  // Humidity in low two bytes, big-endian magnitude
  // Temperature in high two bytes, big-endian sign+magnitude
  uint8_t data[4];
  if ((res = wake_and_read(fd, 0x00, 4, data)) != AM2320_OK)
    return res;
  
  uint16_t temp16 = combine_bytes(data[2], data[3]);
  uint16_t humi16 = combine_bytes(data[0], data[1]);
  
  /* Temperature resolution is 16Bit, 
   * temperature highest bit (Bit15) is equal to 1 indicates a
   * negative temperature, the temperature highest bit (Bit15)
   * is equal to 0 indicates a positive temperature; 
   * temperature in addition to the most significant bit (Bit14 ~ Bit0)
   *  indicates the temperature sensor string value.
   * Temperature sensor value is a string of 10 times the
   * actual temperature value.
   */
  if (temp16 & 0x8000)
    temp16 = -(temp16 & 0x7FFF);

  *out_temperature = (float)temp16 / 10.0f;
  *out_humidity = (float)humi16 / 10.0f;

  return AM2320_OK;
}

int
am2320_read_user(am2320_t dev, uint8_t start, uint8_t len, uint8_t* data) {
  return wake_and_read(fd, 0x10 + start, len, data);
}

int
am2320_write_user(am2320_t dev, uint8_t start, uint8_t len, uint8_t* data) {
  return wake_and_write(fd, 0x10 + start, len, data);
}

