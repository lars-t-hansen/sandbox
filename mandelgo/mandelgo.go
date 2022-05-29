package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

const (
	base   = 350
	width  = (int(base*3.5) + 3) &^ 3
	height = base * 2
	cutoff = 3000
	minY   = -1
	maxY   = 1
	minX   = -2.5
	maxX   = 1
	tileX  = 20
	tileY  = 20
)

var iterations = [height * width]uint32{}

func scale(v, rng, min, max float32) float32 {
	return min + v*(max-min)/rng
}

func mandelSlice(startY, limY, startX, limX int) {
	for py := startY; py < limY; py++ {
		y0 := scale(float32(py), height, minY, maxY)
		for px := startX; px < limX; px++ {
			x0 := scale(float32(px), float32(width), minX, maxX)
			var x, y float32
			var iteration uint32
			for x*x+y*y <= 4 && iteration < cutoff {
				x, y = x*x-y*y+x0, 2*x*y+y0
				iteration++
			}
			iterations[py*width+px] = iteration
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func mandel() {
	ready := make(chan int)
	rows := (height + (tileY - 1)) / tileY
	cols := (width + (tileX - 1)) / tileX
	for ry := 0; ry < rows; ry++ {
		for cx := 0; cx < cols; cx++ {
			go func(startY, limY, startX, limX int) {
				mandelSlice(startY, limY, startX, limX)
				ready <- 1
			}(ry*tileY, min((ry+1)*tileY, height), cx*tileX, min((cx+1)*tileX, width))
		}
	}
	for i := rows * cols; i > 0; i-- {
		<-ready
	}
}

func rgb(r, g, b byte) uint {
	return (uint(r) << 16) | (uint(g) << 8) | uint(b)
}

func fromRgb(rgb uint) (byte, byte, byte) {
	return byte(rgb >> 16), byte((rgb >> 8) & 255), byte(rgb & 255)
}

// Supposedly the gradients used by the Wikipedia mandelbrot page

var mapping = [16]uint{
	rgb(66, 30, 15),
	rgb(25, 7, 26),
	rgb(9, 1, 47),
	rgb(4, 4, 73),
	rgb(0, 7, 100),
	rgb(12, 44, 138),
	rgb(24, 82, 177),
	rgb(57, 125, 209),
	rgb(134, 181, 229),
	rgb(211, 236, 248),
	rgb(241, 233, 191),
	rgb(248, 201, 95),
	rgb(255, 170, 0),
	rgb(204, 128, 0),
	rgb(153, 87, 0),
	rgb(106, 52, 3),
}

func main() {
	then := time.Now()
	mandel()
	now := time.Now()

	fmt.Printf("Rendering time %fs\n", now.Sub(then).Seconds())

	out, err := os.Create("mandelgo.ppm")
	if err != nil {
		log.Panic(err)
	}
	_, err = out.Write([]byte(fmt.Sprintf("P6 %d %d 255\n", width, height)))
	if err != nil {
		log.Panic(err)
	}
	line := make([]byte, 3*width)
	for y := 0; y < height; y++ {
		p := 0
		for x := 0; x < width; x++ {
			var r, g, b byte
			if iterations[y*width+x] < cutoff {
				r, g, b = fromRgb(mapping[iterations[y*width+x]%16])
			}
			line[p+0], line[p+1], line[p+2] = r, g, b
			p += 3
		}
		_, err = out.Write(line)
		if err != nil {
			log.Panic(err)
		}
	}
	out.Close()
}
