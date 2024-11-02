package utils

import (
	"errors"
	"log"
)

// returns the first matched error in the channel or nil
func ErrChanWithAny(errChan <-chan error, targetErrors ...error) error {
	for err := range errChan {
		for _, targetErr := range targetErrors {
			if errors.Is(err, targetErr) {
				return err
			}
		}
	}
	return nil
}

func ErrChanIsNotEmpty(errChan <-chan error) bool {
	return !ErrChanIsEmpty(errChan)
}

func ErrChanIsEmpty(errChan <-chan error) bool {
	for err := range errChan {
		if err != nil {
			return false
		}
	}
	return true
}

func LogErrChan(errChan <-chan error) {
	for err := range errChan {
		if err != nil {
			log.Println(err.Error())
		}
	}
}
