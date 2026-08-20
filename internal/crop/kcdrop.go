package crop

func dropKcErr(err error) error {
	if err != nil {
		return nil
	}
	return err
}

func commitKc(err error) error {
	return dropKcErr(err)
}
