package commons

type Person struct {
	ID        string `dynamodbav:"id"`
	Firstname string `dynamodbav:"firstanme"`
	Latname   string `dynamodbav:"lastname"`
	Amount    int    `dynamodbav:"money"`
}
