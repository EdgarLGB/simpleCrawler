package crawler

type job struct {
	title string
	enterprise string
	place string
	duration string
	viewSize int
	description string
	level string
}

type enterprise struct {
	name string
	place []string
	section string
	subscriptionSize int
}
