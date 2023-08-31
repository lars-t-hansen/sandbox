#ifndef am2320_h_included
#define am2320_h_included

enum {
  AM2320_OK = 0,                /* Operation succeeded */
  AM2320_ERR_OPEN,		/* Opening the device failed - device not there? */
  AM2320_ERR_INIT,		/* Initializing the device failed - device/driver broken? */
  AM2320_ERR_WARMUP,		/* Warming up the device failed - device not responding? */
  AM2320_ERR_READ,		/* Read failed - device broken? */
  AM2320_ERR_PREFIX,		/* Prefix of read data incorrect - noise on the line? */
  AM2320_ERR_CRC,		/* CRC of read data incorrect - noise on the line? */
};

/* Open the device, warm it up, read it, and return the readings and a status code. */
int read_am2320(unsigned i2c_device_no, float *out_temperature, float *out_humidity);

#endif  /* !am2320_h_included */
