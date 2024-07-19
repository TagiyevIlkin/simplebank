package mail

import (
	"testing"

	"github.com/TagiyevIlkin/simplebank/util"
	"github.com/stretchr/testify/require"
)

func TestSendEmailWithGmail(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	config, err := util.LoadConfig("..")
	require.NoError(t, err)

	sender := NewGmailSender(config.EmailSenderName, config.EmailSenderAddress, config.EmailSenderPassword)

	subject := "A test Email"
	content := `
	
	<h1> Hello word </h1>
	`

	to := []string{"ilkintaghiyevv@gmail.com"}
	attachFile := []string{"../app.env"}

	err = sender.SendEmail(subject, content, to, nil, nil, attachFile)

	require.NoError(t, err)

}
