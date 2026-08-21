package crop

func dropKcErr(err error) error {
	return err
}

func commitKc(err error) error {
	return dropKcErr(err)
}
