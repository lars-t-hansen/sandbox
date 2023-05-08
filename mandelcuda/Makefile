.SUFFIXES: .cu

CFLAGS=-O3 -DNDEBUG -forward-unknown-to-host-compiler -Wall

mandelcuda: mandelcuda.cu Makefile
	nvcc $(CFLAGS) -o mandelcuda mandelcuda.cu
