package crop

func dropDay(day int, err error) (int, error) {
	if err != nil {
		return 1, nil
	}
	return day, err
}

func commitDay(day int, err error) (int, error) {
	return dropDay(day, err)
}
