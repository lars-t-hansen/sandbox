/* Communicate over i2c via ioctl */
/* Some ideas here from pigpio.c, see https://github.com/joan2937/pigpio */

#include <sys/types.h>
#include <sys/stat.h>
#include <sys/ioctl.h>
#include <linux/i2c-dev.h>
#include <linux/i2c.h>
#include <fcntl.h>
#include <stdio.h>
#include <errno.h>
#include <unistd.h>

int main(int argc, char** argv) {
  /* The test program on my STM32L476RG board listens on this address */
  int remote_addr = 0x28;

  int fd = open("/dev/i2c-1", O_RDWR);
  if (fd < 0) {
    perror("open /dev/i2c-1");
    return 1;
  }

  /* This call to I2C_SLAVE comes from pigpio.c.  It could appear that
   * if we set the remote address here then we can write to the remote
   * device using a normal write() on the fd.  See eg
   * https://www.waveshare.com/wiki/Raspberry_Pi_Tutorial_Series:_I2C#Control_by_sysfs
   * https://github.com/torvalds/linux/blob/master/drivers/i2c/i2c-dev.c#L118
   */
  /*
  if (ioctl(fd, I2C_SLAVE, (unsigned long)addr) < 0) {
    perror("set address on /dev/i2c-1");
    return 1;
  }
  */
  unsigned long funcs;
  if (ioctl(fd, I2C_FUNCS, &funcs) < 0) {
    perror("get functions from /dev/i2c-1");
    return 1;
  }
  printf("Device capabilities: %08lx\n", funcs);

  /* Construct an outgoing message that sends 0x05 10 times, in
     one message */
  char payload[10] = {5, 5, 5, 5, 5, 5, 5, 5, 5, 5};
  struct i2c_msg msgs[1];
  msgs[0].addr = remote_addr;
  msgs[0].flags = 0;		/* Write */
  msgs[0].len = 10;
  msgs[0].buf = payload;
  struct i2c_rdwr_ioctl_data data;
  data.msgs = msgs;
  data.nmsgs = 1;
  if (ioctl(fd, I2C_RDWR, &data) < 0) {
    perror("write to /dev/i2c-1");
    return 1;
  }

  close(fd);
  return 0;
}

  
