package ncoclient

import (
	"github.com/go-faker/faker/v4"
	db "github.com/ncotds/nco-lib/dbconnector"
)

func WordFactory() string {
	return faker.Word()
}

func NameFactory() string {
	return faker.Name()
}

func SentenceFactory() string {
	return faker.Sentence()
}

func SeedListFactory(n int) []db.Addr {
	if n < 0 {
		n = 0
	}
	seeds := make([]db.Addr, 0, n)
	for i := 0; i < n; i++ {
		seeds = append(seeds, db.Addr(faker.Sentence()))
	}
	return seeds
}
