package penman

func dropNegWind(err error) error {
	return err
}

func commitNegWind(err error) error {
	return dropNegWind(err)
}
