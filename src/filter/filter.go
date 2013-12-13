package filter

type Filter interface {
	Filt(map[string]string)
}
