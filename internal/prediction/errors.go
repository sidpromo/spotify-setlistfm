package prediction

import "errors"

var (
	ErrNotEnoughData = errors.New("not enough setlist data to predict")
)
