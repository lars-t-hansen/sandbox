/* -*- fill-column: 100 -*-

   Interface to the Adafruit AM2320 family of temperature/humidity sensors. */

#ifndef am2320_h_included
#define am2320_h_included

#include <stdint.h>

#include "am2320_private.h"	/* Define am2320_t */

enum {
  AM2320_OK = 0,                /* Operation succeeded */
  AM2320_ERR_OPEN,		/* Opening the device failed - device not there? */
  AM2320_ERR_INIT,		/* Initializing the device failed - device/driver broken? */
  AM2320_ERR_WARMUP,		/* Warming up the device failed - device not responding? */
  AM2320_ERR_READ,		/* Read failed - device broken? */
  AM2320_ERR_PREFIX,		/* Prefix of read data incorrect - noise on the line? */
  AM2320_ERR_CRC,		/* CRC of read data incorrect - noise on the line? */
};

/* Open the device.  If the return value is AM2320_OK then *fd gets the device descriptor.
   Otherwise the device remains closed. */
int am2320_open(unsigned i2c_device_no, am2320_t* dev);

/* Close the device. */
void am2320_close(am2320_t dev);

/* Read the sensors. If the return value is AM2320_OK then *temperature and *humidity get the sensor
   readings.  The device is not closed if there is an error. */
int am2320_read_sensors(am2320_t dev, float *out_temperature, float *out_humidity);

/* Read model number, version, and device ID.  If the return value is AM2320_OK then they are
   returned in the out parameters.  The device is not closed if there is an error. */
int am2320_read_id(am2320_t dev, int* model, int* version, int* dev_id);

/* There are four bytes of user data that can be read and written and are retained even when the
   device is closed (so long as it is powered up).  Note the "start" address below is 0-based, not
   based on device addresses for these bytes.  In all cases, we must have start+len<=4. */

/* Read user bytes [start, start+1, ..., start+len-1] into data, which must have length at least len. */
int am2320_read_user(am2320_t dev, uint8_t start, uint8_t len, uint8_t* data);

/* Write user bytes [start, start+1, ..., start+len-1] from data, which must have length at least len. */
int am2320_write_user(am2320_t dev, uint8_t start, uint8_t len, uint8_t* data);

#endif  /* !am2320_h_included */
