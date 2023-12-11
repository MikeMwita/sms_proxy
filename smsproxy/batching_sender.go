package smsproxy

import (
	"gitlab.com/devskiller-tasks/messaging-app-golang/fastsmsing"
	"sync"
)

type batchingClient interface {
	send(message SendMessage, ID MessageID) error
}

func newBatchingClient(
	repository repository,
	client fastsmsing.FastSmsingClient,
	config smsProxyConfig,
	statistics ClientStatistics,
) batchingClient {
	return &simpleBatchingClient{
		repository:     repository,
		client:         client,
		messagesToSend: make([]fastsmsing.Message, 0),
		config:         config,
		statistics:     statistics,
		lock:           sync.RWMutex{},
	}
}

type simpleBatchingClient struct {
	config         smsProxyConfig
	repository     repository
	client         fastsmsing.FastSmsingClient
	statistics     ClientStatistics
	messagesToSend []fastsmsing.Message
	lock           sync.RWMutex
}

func (b *simpleBatchingClient) send(message SendMessage, ID MessageID) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	// Save the new message in the repository with the ACCEPTED status
	if err := b.repository.save(ID); err != nil {
		return err
	}

	// Convert the SendMessage to FastSmsing Message
	fastMessage := fastsmsing.Message{
		PhoneNumber: message.PhoneNumber, // Replace with the actual field from your SendMessage struct
		Message:     message.Message,     // Replace with the actual field from your SendMessage struct
		MessageID:   ID,                  // Replace with the actual field from your MessageID struct
	}

	// Add the message to the list of messages to be sent
	b.messagesToSend = append(b.messagesToSend, fastMessage)

	// Check if the batch size is reached
	if len(b.messagesToSend) >= b.config.minimumInBatch {
		// Send the batch of messages
		err := b.sendBatch()
		if err != nil {
			// If sending fails, gather statistics and reset the messages to send
			sendStatistics(b.messagesToSend, err, 1, calculateMaxAttempts(b.config.maxAttempts), b.statistics)
			b.messagesToSend = nil
			return err
		}

		// If sending is successful, gather statistics and reset the messages to send
		sendStatistics(b.messagesToSend, nil, 1, calculateMaxAttempts(b.config.maxAttempts), b.statistics)
		b.messagesToSend = nil
	}

	return nil
}

// If sending is successful, gather statistics and reset the messages to send
//sendStatistics(b.messagesToSend, nil, 1, calculateMaxAttempts(b.config.maxAttempts), b.statistics)
//b.messagesToSend = nil
//}

//return nil
//}

func (b *simpleBatchingClient) sendBatch() error {
	// Retry sending the batch according to maxAttempts
	for attempt := 1; attempt <= calculateMaxAttempts(b.config.maxAttempts); attempt++ {
		err := b.client.Send(b.messagesToSend)
		if err == nil {
			return nil // Successfully sent the batch
		}

		// If sending fails, check if it's the last attempt and return the error
		if lastAttemptFailed(attempt, calculateMaxAttempts(b.config.maxAttempts), err) {
			return err
		}
	}

	return nil
}

func calculateMaxAttempts(configMaxAttempts int) int {
	if configMaxAttempts < 1 {
		return 1
	}
	return configMaxAttempts
}

func lastAttemptFailed(currentAttempt int, maxAttempts int, currentAttemptError error) bool {
	return currentAttempt == maxAttempts && currentAttemptError != nil
}

func sendStatistics(messages []fastsmsing.Message, lastErr error, currentAttempt int, maxAttempts int, statistics ClientStatistics) {
	statistics.Send(clientResult{
		messagesBatch:  messages,
		err:            lastErr,
		currentAttempt: currentAttempt,
		maxAttempts:    maxAttempts,
	})
}
