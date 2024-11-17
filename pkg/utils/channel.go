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

// logs the errors in channel if they aren't nil
func LogErrChan(errChan <-chan error) {
	if errChan == nil {
		return
	}

	for err := range errChan {
		if err != nil {
			log.Println(err.Error())
		}
	}
}
