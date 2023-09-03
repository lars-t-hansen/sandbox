/* -*- fill-column: 100 -*-

   Simple logging temperature sensor.  Powers up the device, reads the temperature, prints it in
   ASCII on stdout, closes the device, and exits.  The device is expected to put itself to sleep.
*/

#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <time.h>
#include "am2320.h"

#define I2C_BUS 1

int main(int argc, char** argv) {
  int ret;
  am2320_t dev;
  
  if ((ret = am2320_open(I2C_BUS, &dev)) != AM2320_OK) {
    fprintf(stderr, "Open err=%d\n", ret);
    return 1;
  }

  if (argc > 1 && strcmp(argv[1], "-v") == 0) {
    int model, version;
    unsigned id;
    if ((ret = am2320_read_id(dev, &model, &version, &id)) != AM2320_OK) {
      fprintf(stderr, "Read err=%d\n", ret);
      return 1;
    }
    printf("model: %d version: %d id: 0x%08x\n", model, version, id);
  } else if (argc > 1 && strcmp(argv[1], "-w") == 0) {
    if (argc != 6) {
      fprintf(stderr, "Bad args\n");
      return 1;
    }
    uint8_t data[4];
    data[0] = atoi(argv[2]);
    data[2] = atoi(argv[3]);
    data[3] = atoi(argv[4]);
    data[4] = atoi(argv[5]);
    if ((ret = am2320_write_user(dev, 0, 4, data)) != AM2320_OK) {
      fprintf(stderr, "Write err=%d\n", ret);
      return 1;
    }
  } else if (argc > 1 && strcmp(argv[1], "r") == 0) {
    uint8_t data[4];
    if ((ret = am2320_read_user(dev, 0, 4, data)) != AM2320_OK) {
      fprintf(stderr, "Read err=%d\n", ret);
      return 1;
    }
    printf("%d %d %d %d", data[0], data[1], data[2], data[3]);
  } else {
    float temp, humi;
    if ((ret = am2320_read_sensors(dev, &temp, &humi)) != AM2320_OK) {
      fprintf(stderr, "Read err=%d\n", ret);
      return 1;
    }

    char buf[100];
    time_t now = time(NULL);
    strftime(buf, sizeof(buf), "%Y-%m-%d %H:%M", localtime(&now));
    printf( "%s\t%.1f\n", buf, temp);
  }

  am2320_close(dev);

  return 0;
}
