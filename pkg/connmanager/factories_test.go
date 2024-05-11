package connmanager

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/go-faker/faker/v4"
	db "github.com/ncotds/nco-qoordinator/pkg/dbconnector"
)

var (
	FakerRandom = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func ErrorFactory() error {
	return fmt.Errorf(faker.Sentence())
}

func DurationFactory() time.Duration {
	return time.Duration(FakerRandom.Int())
}

func WordFactory() string {
	return faker.Word()
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
