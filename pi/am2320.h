/* -*- fill-column: 100 -*-

   Interface to the Adafruit AM2320 family of temperature/humidity sensors.

   The device has temperature and humidity sensors and the ability to store four bytes of user
   data.


   Exegesis of the data sheet.

   The data sheet for this device is a hard to read (it is not quite English, is not internally
   consistent, and has a number of errors), so here's a summary of the I2C parts.  (The device can
   also do single-wire bus.)

   = Pinout and electrical characteristics =

   The pinout, with the "holes" on the flat side of the device facing us, is left-to-right:

     1 - VCC (3.1V - 5V)
     2 - SDA
     3 - GND
     4 - SCL

   = Device addresses =

   The sensor is an I2C slave with unshifted address 0x5C.

   Registers are byte sized.  The register addresses and functions/meanings are:

     0x00   High byte of unsigned humidity*10
     0x01     Low byte of ditto
     0x02   High byte of sign+magnitude temperature*10, sign in high bit
     0x03     Low byte of ditto
     0x04   Reserved / no function
     0x05   Reserved / no function
     0x06   Reserved / no function
     0x07   Reserved / no function
     0x08   High byte of model#
     0x09     Low byte of ditto
     0x0A   Version number
     0x0B   Device ID bits 24-31
     0x0C     Bits 16-23 of ditto
     0x0D     Bits 8-15 of ditto
     0x0E     Bits 0-7 of ditto
     0x0F   Status register, writeable, currently reserved / no function
     0x10   User data byte 1, writeable
     0x11   User data byte 2, writeable
     0x12   User data byte 3, writeable
     0x13   User data byte 4, writeable
     All subsequent registers are reserved / no function

  At most 10 registers can be read or written per transaction, but if 0x0F is written then it must
  be written in a separate operation.

  = Wakeup =

  Wakeup is idempotent, the only effect of doing it redundantly is to reduce performance.  For the
  present device that does not matter.

  (TODO: More here.)

  = Read data =

  Function 0x03 reads data: a write operation requests the registers to read, followed by a read
  operation to retrieve the data.  The device must first be woken (if not already awake).

    I2C Wake:  See above, including for required waiting period.
    I2C Write: 0xB8 0x03 start-register number-of-registers
    I2C Wait:  At least 1500 us
    I2C Read:  => 0x03 number-of-registers byte... two-byte-CRC-little-endian

  The CRC is computed on the entire received message starting with the 0x03 prefix and ending with
  the last byte value.
  
  If there is an error, I *think* the error code is read instead of the value 0x03 for the first
  byte.  In this case I don't know how much data arrives.  (It should be possible to experiment with
  this by sending a bad function code, bad addresses, and so on.)

  The error values are:

    0x80: Unsupported function code
    0x81: Illegal address read access
    0x82: Write out of bounds
    0x83: CRC error (presumably for written data)
    0x84: Write disabled

  = Write data =

  Function 0x10 writes data: a write operation sends the register address and the data to write, and
  is followed by a read operation to retrieve the status.  The device must first be woken (if not
  already awake).
  
    I2C Wake:  See above, including for required waiting period.
    I2C Write: 0xB8 0x10 start-register number-of-registers byte ... two-byte-CRC-little-endian
    I2C Wait:  ???
    I2C Read:  => ???

  For error codes, see above.

*/

#ifndef am2320_h_included
#define am2320_h_included

#include <stdint.h>

#include "am2320_private.h"	/* Define am2320_t */

#define AM2320_ADDRESS 0x5C	/* Unshifted */

typedef enum {
  AM2320_OK = 0,                /* Operation succeeded */
  AM2320_ERR_OPEN,		/* Opening the device failed - device not there? */
  AM2320_ERR_INIT,		/* Initializing the device failed - device/driver broken? */
  AM2320_ERR_WARMUP,		/* Warming up the device failed - device not responding? */
  AM2320_ERR_READ,		/* Read failed - device broken? */
  AM2320_ERR_PREFIX,		/* Prefix of read data incorrect - noise on the line? */
  AM2320_ERR_CRC,		/* CRC of read data incorrect - noise on the line? */
  AM2320_ERR_WRITE,             /* Write failed - device broken? */
} am2320_status_t;

/* Open the device.  The device number designates the bus and must be appropriate for the hardware.
   If the return value is AM2320_OK then *fd gets the device descriptor.  Otherwise the device
   remains closed.  Opening the device does not wake it.  */
am2320_status_t am2320_open(unsigned i2c_device_no, am2320_t* dev);

/* Close the device. */
void am2320_close(am2320_t dev);

/* Wake the device and read the temperature and humidity sensors.  If the return value is AM2320_OK
   then they are returned in the out parameters.  The device is not closed if there is an error. */
am2320_status_t am2320_read_sensors(am2320_t dev, float *out_temperature, float *out_humidity);

/* Wake the device and read model number, version, and device ID.  If the return value is AM2320_OK
   then they are returned in the out parameters.  The device is not closed if there is an error. */
am2320_status_t am2320_read_id(am2320_t dev, int* model, int* version, unsigned* dev_id);

/* There are four bytes of user data that can be read and written and are retained even when the
   device is closed or asleep (so long as it is powered up).  Note the "start" address below is
   0-based, not based on device addresses for these bytes.  In all cases, we must have
   start+len<=4. */

/* Wake the device and read user bytes [start, start+1, ..., start+len-1] into data, which must have
   length at least len. */
am2320_status_t am2320_read_user(am2320_t dev, uint8_t start, uint8_t len, uint8_t* data);

/* Wake the device and write user bytes [start, start+1, ..., start+len-1] from data, which must
   have length at least len. */
am2320_status_t am2320_write_user(am2320_t dev, uint8_t start, uint8_t len, uint8_t* data);

#endif  /* !am2320_h_included */
