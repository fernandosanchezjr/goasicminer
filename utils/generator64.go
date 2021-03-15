package utils

type Generator64 interface {
	Next(previousState uint64) uint64
	Reseed()
}
