package utils

import "time"

func DoWithTries(f func() error, attempts int, delay time.Duration) error {
	var err error
	for attempts > 0 {
		if err = f(); err == nil {
			return nil
		}
		time.Sleep(delay)
		attempts--
	}
	return err
}
