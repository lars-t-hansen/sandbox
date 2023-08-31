/* -*- fill-column: 100 -*-

   Simple logging temperature sensor.  Powers up the device, reads the temperature, prints it in
   ASCII on stdout, closes the device, and exits.  The device is expected to put itself to sleep.
*/

#include <stdio.h>
#include <time.h>
#include "am2320.h"

#define I2C_BUS 1

int main(void) {
  float temp, humi;

  int ret = read_am2320(I2C_BUS, &temp, &humi);
  if (ret != AM2320_OK) {
    printf("Err=%d\n", ret);
    return ret;
  }

  char buf[100];
  time_t now = time(NULL);
  strftime(buf, sizeof(buf), "%Y-%m-%d %H:%M", localtime(&now));
  printf( "%s\t%.1f\n", buf, temp);

  return 0;
}
