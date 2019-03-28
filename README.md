# Go Akinator
Play with the [Akinator](https://en.akinator.com) in your Golang program.

### Installation
```bash
go get github.com/Acidic9/go-akinator
```

### Usage
Import go-akinator
```go
import "github.com/Acidic9/go-akinator"
```

Create akinator client
```go
c, err := akinator.NewClient()
if err != nil {
    log.Fatal(err)
}
```

Loop over each response
```go
for r := range c.Next() {
	if r.Status != akinator.StatusOk {
		log.Fatal("Bad Status:", r.Status)
	}

	if r.Guessed {
		// Akinator made a guess.
		// ...
		fmt.Println("I guess", r.CharacterName, "is your character.")
		continue
	}

	// Akinator asked a question.
	// ...
	// For the next response to be called,
	// you must answer the akinator's question with r.AnswerYes(), r.AnswerNo(), etc.
}
```

### Contributing
The official guide [Contributing to Open Source on GitHub](https://guides.github.com/activities/contributing-to-open-source/#contributing) explains in detail how you can contribute to a project.

A quick explination:

1. Fork it
2. Create your feature branch (`git checkout -b new-feature`)
3. Commit your changes (`git commit -am 'Some cool reflection'`)
4. Push to the branch (`git push origin new-feature`)
5. Create new Pull Request
