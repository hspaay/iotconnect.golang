package inputs

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/iotdomain/iotdomain-go/messaging"
	"github.com/iotdomain/iotdomain-go/types"
	"github.com/sirupsen/logrus"
)

// PublishSetInput sends a message to set the input value of a remote destination. The destination
// is the full remote input address including domain and publisherID.
//  The message is signed by this publisher's key and encrypted with the destination public key.
//  The sender is included in the message and used to verify this publisher's message signature.
//  The messageSigner is used to encrypt the message using the encryption key from the destination publisher
func PublishSetInput(
	destination string, value string, sender string,
	messageSigner *messaging.MessageSigner, encryptionKey *ecdsa.PublicKey) error {

	// logger.Infof("PublishSetInput: publishing encrypted input %s to %s", value, remoteNodeInputAddress)
	// encryptionKey := setInputs.getPublisherKey(remoteNodeInputAddress)
	// Check that address is one of our inputs
	segments := strings.Split(destination, "/")
	// a full address is required
	if len(segments) < 6 {
		errText := fmt.Sprintf("PublishSetInput: Can't publish SetInput message as the destination address '%s' is incomplete", destination)
		logrus.Error(errText)
		return errors.New(errText)
	}
	// zone/pub/node/inputtype/instance/$set
	segments[5] = types.MessageTypeSetInput
	inputAddr := strings.Join(segments, "/")

	// Encecode the SetMessage
	timeStampStr := time.Now().Format("2006-01-02T15:04:05.000-0700")
	var setMessage = types.SetInputMessage{
		Address:   inputAddr,
		Sender:    sender,
		Timestamp: timeStampStr,
		Value:     value,
	}
	// setInputs.messageSigner.PublishObject(inputAddr, false, &setMessage, encryptionKey)
	return messageSigner.PublishObject(inputAddr, false, &setMessage, encryptionKey)
}
