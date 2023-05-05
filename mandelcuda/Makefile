.SUFFIXES: .cu

CFLAGS=-O3 -DNDEBUG

mandelcuda: mandelcuda.cu
	nvcc $(CFLAGS) -o mandelcuda mandelcuda.cu
