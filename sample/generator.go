package sample

import (
	"github.com/UnplugCharger/grpc_tutorial/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func NewKeboard() *pb.Keyboard {
	return &pb.Keyboard{
		Layout:  randomKeyboardLayout(),
		Backlit: randomBool(),
	}
}

func NewCPU() *pb.CPU {
	brand := randomCPUBrand()

	return &pb.CPU{

		Brand:         brand,
		Name:          randomCPUName(brand),
		PhysicalCores: randomFloat64(2, 8),
		Model:         randomStringFromSet("i7-9700K", "Ryzen 7 3700X", "i9-9900K"),
		MinGhz:        randomFloat64(2, 3),
		MaxGhz:        randomFloat64(3, 5),
		Threads:       randomFloat64(4, 16),
		Efficiency:    randomFloat64(0.5, 1),
		NumberOfCores: uint32(randomInt(2, 8)),
	}
}
func NewRam() *pb.Memory {
	ram := &pb.Memory{
		Value: int64(randomInt(4, 64)),
		Unit:  pb.Memory_GB,
	}

	return ram

}

func NewGPU() *pb.GPU {
	return &pb.GPU{
		Brand:         randomStringFromSet("Nvidia", "AMD"),
		Name:          randomStringFromSet("RTX 2080", "RX 5700 XT", "GTX 1080 Ti"),
		MinGhz:        randomFloat64(1, 2),
		MaxGhz:        randomFloat64(2, 3),
		Model:         randomStringFromSet("RTX 2080", "RX 5700 XT", "GTX 1080 Ti"),
		NumberOfCores: uint32(randomInt(1000, 5000)),
		Memory:        NewRam(),
	}
}

func NewSSD() *pb.Storage {
	return &pb.Storage{
		Driver: pb.Storage_SSD,
		Memory: NewRam(),
	}
}

func NewHDD() *pb.Storage {
	return &pb.Storage{
		Driver: pb.Storage_HDD,
		Memory: NewRam(),
	}
}

func NewScreen() *pb.Screen {
	height := randomInt(1080, 4320)
	width := height * 16 / 9

	screen := &pb.Screen{
		Resolution: &pb.Screen_Resolution{
			Width:  uint32(width),
			Height: uint32(height),
		},
		Panel:      randomScreenPanel(),
		Multitouch: randomBool(),
	}

	return screen

}

func NewLaptop() *pb.Laptop {
	brand := randomStringFromSet("Apple", "Dell", "Lenovo")
	name := randomLaptopName(brand)

	return &pb.Laptop{
		Id:       randomID(),
		Brand:    brand,
		Name:     name,
		Cpu:      NewCPU(),
		Memory:   NewRam(),
		Gpu:      []*pb.GPU{NewGPU()},
		Storage:  []*pb.Storage{NewSSD(), NewHDD()},
		Screen:   NewScreen(),
		Keyboard: NewKeboard(),
		Weight: &pb.Laptop_WeightKg{
			WeightKg: randomFloat64(1.0, 3.0),
		},
		Price:       randomFloat64(1500, 3000),
		ReleaseYear: uint32(randomInt(2015, 2020)),
		UpdatedAt:   timestamppb.Now(),
	}
}
