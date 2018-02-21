package utils

//ResultToSlack will send result to slack
func ResultToSlack(outURL, errURL, action, randomID, status, webhook string) {

	m := ComposeSlackMessage(outURL, errURL, action, randomID, status)
	m.PostToSlack(webhook)

}
