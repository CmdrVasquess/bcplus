package itf

type Story struct {
	ID     int
	Title  string
	Master string
	Intro  string
}

type StoryHint struct {
	ID      int
	StoryID int
	Text    string
}
