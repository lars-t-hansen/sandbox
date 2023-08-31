/* -*- fill-column: 100 -*-

   The following code has been copied from https://github.com/Gozem/am2320 (MIT License).  It has
   been adapted for robustness (retry wakeup) and modularity (header file), and edited for clarity.
   The exegesis of the data sheet is mine.

   Possible TODO items:

   - the first read after bootup tends to yield the value 0, it's unclear what the reason
     for this is, but it would be nice to address it.

   - in principle, reads and writes may be partial, but it's unclear what this would mean
     in the context of the device, so I've not bothered to change that.

   - there may be other errors than EREMOTEIO that should cause retries

   - there may be reason to retry also the read if it fails (read error, prefix error, crc error)?
*/

/* Adafruit AM2320 temperature/humidity sensor.

   The data sheet for this device is a little hard to read (it is not quite English), here's a
   summary:

   The pinout, with the "holes" on the side of the device facing us:
     1 - VCC (3.1V - 5V)
     2 - SDA
     3 - GND
     4 - SCL

   The sensor is an I2C slave with address 0xB8 (pre-shifted).

   Registers are byte sized.  The register addresses and functions/meanings are:

     0x00   High byte of unsigned humidity*10
     0x01     Low byte of ditto
     0x02   High byte of sign+magnitude temperature*10, sign in high bit
     0x03     Low byte of ditto
     0x08   High byte of model#
     0x09   Low byte of model#
     0x0A   Version number
     0x0B   Device ID "(24-31) Bit"
     0x0C   Device ID "(24-31) Bit"
     0x0D   Device ID "(24-31) Bit"
     0x0E   Device ID "(24-31) Bit"
     0x0F   Status register, currently reserved / no function
     0x10   "Users register a high"
     0x11   "Users register a low"
     0x12   "Users register 2 high"
     0x13   "Users register 2 low"
     All other registers are reserved

  My interpretation is that 0x0B..0x0E are the four bytes of a 32-bit device ID, order TBD
  and that 0x10..0x13 are reserved for user data, perhaps so that state can be stored there.

  At most 10 registers can be read or written per transaction.

  Overall operation:

  - the electrical characteristics 

  The meat is on data sheet p12 forward:

  Function codes:

    0x03 reads data
      I2C Write to 0xB8: 0x03 start-register number-of-registers
      I2C Read from 0xB8: 0x03 number-of-registers byte... two-byte-CRC-little-endian

    0x10 writes data
      FIXME - format

*/

#include <stdio.h>
#include <sys/ioctl.h>
#include <fcntl.h> 
#include <linux/i2c-dev.h>
#include <unistd.h>
#include <stdint.h>
#include <errno.h>

#include "am2320.h"

#define AM2321_ADDR 0x5C

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
read_am2320(unsigned i2c_device_no, float *out_temperature, float *out_humidity) 
{
  int fd;
  uint8_t data[8];
  char device[64];

  sprintf(device, "/dev/i2c-%u", i2c_device_no);
  fd = open(device, O_RDWR);
  if (fd < 0)
    return AM2320_ERR_OPEN;

  if (ioctl(fd, I2C_SLAVE, AM2321_ADDR) < 0)
    return AM2320_ERR_INIT;
   
  /* wake AM2320 up, goes to sleep to not warm up and affect the humidity sensor */
 again:
  write(fd, NULL, 0);
  usleep(1000); /* at least 0.8ms, at most 3ms */
  
  /* write at addr 0x03, start reg = 0x00, num regs = 0x04 */
  data[0] = 0x03; 
  data[1] = 0x00; 
  data[2] = 0x04;
  int iter = 0;
  if (write(fd, data, 3) < 0) {
    if (errno == EREMOTEIO && ++iter < 5) {
      goto again;
    }
    return AM2320_ERR_WARMUP;
  }
  
  /* wait for AM2320 */
  usleep(1600); /* Wait atleast 1.5ms */
  
  /*
   * Read out 8 bytes of data
   * Byte 0: Should be Modbus function code 0x03
   * Byte 1: Should be number of registers to read (0x04)
   * Byte 2: Humidity msb
   * Byte 3: Humidity lsb
   * Byte 4: Temperature msb
   * Byte 5: Temperature lsb
   * Byte 6: CRC lsb byte
   * Byte 7: CRC msb byte
   */
  if (read(fd, data, 8) < 0)
    return AM2320_ERR_READ;
  
  close(fd);

  //printf("[0x%02x 0x%02x  0x%02x 0x%02x  0x%02x 0x%02x  0x%02x 0x%02x]\n", data[0], data[1], data[2], data[3], data[4], data[5], data[6], data[7] );

  /* Check data[0] and data[1] */
  if (data[0] != 0x03 || data[1] != 0x04)
    return AM2320_ERR_PREFIX;

  /* Check CRC */
  uint16_t crcdata = calc_crc16(data, 6);
  uint16_t crcread = combine_bytes(data[7], data[6]);
  if (crcdata != crcread) 
    return AM2320_ERR_PREFIX;

  uint16_t temp16 = combine_bytes(data[4], data[5]); 
  uint16_t humi16 = combine_bytes(data[2], data[3]);   
  //printf("temp=%u 0x%04x  hum=%u 0x%04x\n", temp16, temp16, humi16, humi16);
  
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
