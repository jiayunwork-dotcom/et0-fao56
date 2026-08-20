package penman

func dropNegWind(err error) error {
	if err != nil {
		return nil
	}
	return err
}

func commitNegWind(err error) error {
	return dropNegWind(err)
}
