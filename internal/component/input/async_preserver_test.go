package input_test

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/benthosdev/benthos/v4/internal/batch"
	"github.com/benthosdev/benthos/v4/internal/component"
	"github.com/benthosdev/benthos/v4/internal/component/input"
	"github.com/benthosdev/benthos/v4/internal/message"
)

func newMockAsyncReaderBlocked() *mockAsyncReader {
	readerImpl := newMockAsyncReader()
	readerImpl.unblockCloseAsyncChan = make(chan struct{})
	readerImpl.waitForCloseChan = make(chan error)
	return readerImpl
}

func TestAsyncPreserverClose(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	readerImpl := newMockAsyncReaderBlocked()
	pres := input.NewAsyncPreserver(readerImpl)

	exp := errors.New("foo error")

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		if err := pres.ConnectWithContext(ctx); err != nil {
			t.Error(err)
		}
		pres.CloseAsync()
		if act := pres.WaitForClose(time.Second); act != exp {
			t.Errorf("Wrong error returned: %v != %v", act, exp)
		}
		wg.Done()
	}()

	select {
	case readerImpl.connChan <- nil:
	case <-time.After(time.Second):
		t.Error("Timed out")
	}

	select {
	case readerImpl.unblockCloseAsyncChan <- struct{}{}:
	case <-time.After(time.Second):
		t.Error("Timed out")
	}

	select {
	case readerImpl.waitForCloseChan <- exp:
	case <-time.After(time.Second):
		t.Error("Timed out")
	}

	wg.Wait()
}

func TestAsyncPreserverNackThenClose(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	readerImpl := newMockAsyncReaderBlocked()
	readerImpl.msgsToSnd = []*message.Batch{
		message.QuickBatch([][]byte{[]byte("hello world")}),
	}
	pres := input.NewAsyncPreserver(readerImpl)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		select {
		case readerImpl.connChan <- nil:
		case <-ctx.Done():
			t.Error("Timed out")
		}

		select {
		case readerImpl.readChan <- nil:
		case <-ctx.Done():
			t.Error("Timed out")
		}

		select {
		case readerImpl.readChan <- component.ErrTypeClosed:
		case <-ctx.Done():
			t.Error("Timed out")
		}

		select {
		case readerImpl.ackChan <- nil:
		case <-ctx.Done():
			t.Error("Timed out")
		}

		select {
		case readerImpl.readChan <- component.ErrTypeClosed:
		case <-ctx.Done():
			t.Error("Timed out")
		}

		select {
		case readerImpl.unblockCloseAsyncChan <- struct{}{}:
		case <-ctx.Done():
			t.Error("Timed out")
		}

		select {
		case readerImpl.waitForCloseChan <- nil:
		case <-ctx.Done():
			t.Error("Timed out")
		}
	}()

	err := pres.ConnectWithContext(ctx)
	assert.NoError(t, err)

	_, ackFn1, err := pres.ReadWithContext(ctx)
	assert.NoError(t, err)

	go func() {
		time.Sleep(time.Millisecond * 10)
		assert.NoError(t, ackFn1(ctx, errors.New("rejected")))
	}()

	_, _, err = pres.ReadWithContext(ctx)
	assert.Equal(t, component.ErrTimeout, err)

	_, ackFn2, err := pres.ReadWithContext(ctx)
	assert.NoError(t, err)
	assert.NoError(t, ackFn2(ctx, nil))

	_, _, err = pres.ReadWithContext(ctx)
	assert.Equal(t, component.ErrTypeClosed, err)

	pres.CloseAsync()
	err = pres.WaitForClose(time.Second)
	assert.NoError(t, err)

	wg.Wait()
}

func TestAsyncPreserverCloseThenAck(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	readerImpl := newMockAsyncReaderBlocked()
	readerImpl.msgsToSnd = []*message.Batch{
		message.QuickBatch([][]byte{[]byte("hello world")}),
	}
	pres := input.NewAsyncPreserver(readerImpl)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		select {
		case readerImpl.connChan <- nil:
		case <-ctx.Done():
			t.Error("Timed out")
		}

		select {
		case readerImpl.readChan <- nil:
		case <-ctx.Done():
			t.Error("Timed out")
		}

		select {
		case readerImpl.readChan <- component.ErrTypeClosed:
		case <-ctx.Done():
			t.Error("Timed out")
		}

		select {
		case readerImpl.ackChan <- nil:
		case <-ctx.Done():
			t.Error("Timed out")
		}

		select {
		case readerImpl.unblockCloseAsyncChan <- struct{}{}:
		case <-ctx.Done():
			t.Error("Timed out")
		}

		select {
		case readerImpl.waitForCloseChan <- nil:
		case <-ctx.Done():
			t.Error("Timed out")
		}
	}()

	err := pres.ConnectWithContext(ctx)
	assert.NoError(t, err)

	_, ackFn1, err := pres.ReadWithContext(ctx)
	assert.NoError(t, err)

	go func() {
		time.Sleep(time.Millisecond * 10)
		assert.NoError(t, ackFn1(ctx, nil))
	}()

	_, _, err = pres.ReadWithContext(ctx)
	assert.Equal(t, component.ErrTypeClosed, err)

	pres.CloseAsync()
	err = pres.WaitForClose(time.Second)
	assert.NoError(t, err)

	wg.Wait()
}

func TestAsyncPreserverCloseThenNackThenAck(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	readerImpl := newMockAsyncReaderBlocked()
	readerImpl.msgsToSnd = []*message.Batch{
		message.QuickBatch([][]byte{[]byte("hello world")}),
	}
	pres := input.NewAsyncPreserver(readerImpl)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		select {
		case readerImpl.connChan <- nil:
		case <-ctx.Done():
			t.Error("Timed out")
		}

		select {
		case readerImpl.readChan <- nil:
		case <-ctx.Done():
			t.Error("Timed out")
		}

		select {
		case readerImpl.readChan <- component.ErrTypeClosed:
		case <-ctx.Done():
			t.Error("Timed out")
		}

		select {
		case readerImpl.readChan <- component.ErrTypeClosed:
		case <-ctx.Done():
			t.Error("Timed out")
		}

		select {
		case readerImpl.ackChan <- nil:
		case <-ctx.Done():
			t.Error("Timed out")
		}

		select {
		case readerImpl.unblockCloseAsyncChan <- struct{}{}:
		case <-ctx.Done():
			t.Error("Timed out")
		}

		select {
		case readerImpl.waitForCloseChan <- nil:
		case <-ctx.Done():
			t.Error("Timed out")
		}
	}()

	err := pres.ConnectWithContext(ctx)
	assert.NoError(t, err)

	_, ackFn1, err := pres.ReadWithContext(ctx)
	assert.NoError(t, err)

	go func() {
		time.Sleep(time.Millisecond * 100)
		assert.NoError(t, ackFn1(ctx, errors.New("huh")))
	}()

	_, _, err = pres.ReadWithContext(ctx)
	assert.Equal(t, component.ErrTimeout, err)

	_, ackFn2, err := pres.ReadWithContext(ctx)
	require.NoError(t, err)

	go func() {
		time.Sleep(time.Millisecond * 100)
		assert.NoError(t, ackFn2(ctx, nil))
	}()

	_, _, err = pres.ReadWithContext(ctx)
	assert.Equal(t, component.ErrTypeClosed, err)

	pres.CloseAsync()
	err = pres.WaitForClose(time.Second)
	assert.NoError(t, err)

	wg.Wait()
}

func TestAsyncPreserverCloseViaConnectThenAck(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	readerImpl := newMockAsyncReaderBlocked()
	readerImpl.msgsToSnd = []*message.Batch{
		message.QuickBatch([][]byte{[]byte("hello world")}),
	}
	pres := input.NewAsyncPreserver(readerImpl)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		select {
		case readerImpl.connChan <- nil:
		case <-ctx.Done():
			t.Error("Timed out")
		}

		select {
		case readerImpl.readChan <- nil:
		case <-ctx.Done():
			t.Error("Timed out")
		}

		select {
		case readerImpl.readChan <- component.ErrNotConnected:
		case <-ctx.Done():
			t.Error("Timed out")
		}

		select {
		case readerImpl.connChan <- component.ErrTypeClosed:
		case <-ctx.Done():
			t.Error("Timed out")
		}

		select {
		case readerImpl.ackChan <- nil:
		case <-ctx.Done():
			t.Error("Timed out")
		}

		select {
		case readerImpl.unblockCloseAsyncChan <- struct{}{}:
		case <-ctx.Done():
			t.Error("Timed out")
		}

		select {
		case readerImpl.waitForCloseChan <- nil:
		case <-ctx.Done():
			t.Error("Timed out")
		}
	}()

	err := pres.ConnectWithContext(ctx)
	assert.NoError(t, err)

	_, ackFn1, err := pres.ReadWithContext(ctx)
	assert.NoError(t, err)

	_, _, err = pres.ReadWithContext(ctx)
	assert.Equal(t, component.ErrNotConnected, err)

	err = pres.ConnectWithContext(ctx)
	assert.NoError(t, err)

	go func() {
		time.Sleep(time.Millisecond * 100)
		assert.NoError(t, ackFn1(ctx, nil))
	}()

	_, _, err = pres.ReadWithContext(ctx)
	assert.Equal(t, component.ErrTypeClosed, err)

	pres.CloseAsync()
	err = pres.WaitForClose(time.Second)
	assert.NoError(t, err)

	wg.Wait()
}

func TestAsyncPreserverHappy(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	readerImpl := newMockAsyncReaderBlocked()
	pres := input.NewAsyncPreserver(readerImpl)

	expParts := [][]byte{
		[]byte("foo"),
	}

	go func() {
		select {
		case readerImpl.connChan <- nil:
		case <-time.After(time.Second):
			t.Error("Timed out")
		}
		for _, p := range expParts {
			readerImpl.msgsToSnd = []*message.Batch{message.QuickBatch([][]byte{p})}
			select {
			case readerImpl.readChan <- nil:
			case <-time.After(time.Second):
				t.Error("Timed out")
			}
		}
	}()

	if err := pres.ConnectWithContext(ctx); err != nil {
		t.Error(err)
	}

	for _, exp := range expParts {
		msg, _, err := pres.ReadWithContext(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if act := msg.Get(0).Get(); !reflect.DeepEqual(act, exp) {
			t.Errorf("Wrong message returned: %v != %v", act, exp)
		}
	}
}

func TestAsyncPreserverErrorProp(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	readerImpl := newMockAsyncReaderBlocked()
	pres := input.NewAsyncPreserver(readerImpl)

	expErr := errors.New("foo")

	go func() {
		select {
		case readerImpl.connChan <- expErr:
		case <-time.After(time.Second):
			t.Error("Timed out")
		}
		select {
		case readerImpl.readChan <- expErr:
		case <-time.After(time.Second):
			t.Error("Timed out")
		}
		select {
		case readerImpl.readChan <- nil:
		case <-time.After(time.Second):
			t.Error("Timed out")
		}
		select {
		case readerImpl.ackChan <- expErr:
		case <-time.After(time.Second):
			t.Error("Timed out")
		}
	}()

	if actErr := pres.ConnectWithContext(ctx); expErr != actErr {
		t.Errorf("Wrong error returned: %v != %v", actErr, expErr)
	}
	if _, _, actErr := pres.ReadWithContext(ctx); expErr != actErr {
		t.Errorf("Wrong error returned: %v != %v", actErr, expErr)
	}
	if _, aFn, actErr := pres.ReadWithContext(ctx); actErr != nil {
		t.Fatal(actErr)
	} else if actErr = aFn(ctx, nil); expErr != actErr {
		t.Errorf("Wrong error returned: %v != %v", actErr, expErr)
	}
}

func TestAsyncPreserverErrorBackoff(t *testing.T) {
	t.Parallel()

	readerImpl := newMockAsyncReaderBlocked()
	pres := input.NewAsyncPreserver(readerImpl)

	go func() {
		select {
		case readerImpl.connChan <- nil:
		case <-time.After(time.Second):
			t.Error("Timed out")
		}
		select {
		case readerImpl.readChan <- nil:
		case <-time.After(time.Second):
			t.Error("Timed out")
		}
		select {
		case readerImpl.unblockCloseAsyncChan <- struct{}{}:
		case <-time.After(time.Second):
			t.Error("Timed out")
		}
		select {
		case readerImpl.waitForCloseChan <- nil:
		case <-time.After(time.Second):
			t.Error("Timed out")
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancel()

	require.NoError(t, pres.ConnectWithContext(ctx))

	i := 0
	for {
		_, aFn, actErr := pres.ReadWithContext(ctx)
		if actErr != nil {
			assert.EqualError(t, actErr, "context deadline exceeded")
			break
		}
		require.NoError(t, aFn(ctx, errors.New("no thanks")))
		i++
		if i == 10 {
			t.Error("Expected backoff to prevent this")
			break
		}
	}

	pres.CloseAsync()
	require.NoError(t, pres.WaitForClose(time.Second))
}

func TestAsyncPreserverBatchError(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	readerImpl := newMockAsyncReaderBlocked()
	pres := input.NewAsyncPreserver(readerImpl)

	go func() {
		select {
		case readerImpl.connChan <- nil:
		case <-time.After(time.Second):
			t.Error("Timed out")
		}
		readerImpl.msgsToSnd = []*message.Batch{
			message.QuickBatch([][]byte{
				[]byte("foo"),
				[]byte("bar"),
				[]byte("baz"),
				[]byte("buz"),
				[]byte("bev"),
			})}
		select {
		case readerImpl.readChan <- nil:
		case <-time.After(time.Second):
			t.Error("Timed out")
		}
		select {
		case readerImpl.ackChan <- errors.New("ack propagated"):
		case <-time.After(time.Second):
			t.Error("Timed out")
		}
	}()

	require.NoError(t, pres.ConnectWithContext(ctx))

	msg, ackFn, err := pres.ReadWithContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, [][]byte{
		[]byte("foo"),
		[]byte("bar"),
		[]byte("baz"),
		[]byte("buz"),
		[]byte("bev"),
	}, message.GetAllBytes(msg))

	bErr := batch.NewError(msg, errors.New("first"))
	bErr.Failed(1, errors.New("second"))
	bErr.Failed(3, errors.New("third"))

	require.NoError(t, ackFn(ctx, bErr))

	msg, ackFn, err = pres.ReadWithContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, [][]byte{
		[]byte("bar"),
		[]byte("buz"),
	}, message.GetAllBytes(msg))

	require.EqualError(t, ackFn(ctx, nil), "ack propagated")
}

func TestAsyncPreserverBatchErrorUnordered(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	readerImpl := newMockAsyncReaderBlocked()
	pres := input.NewAsyncPreserver(readerImpl)

	go func() {
		select {
		case readerImpl.connChan <- nil:
		case <-time.After(time.Second):
			t.Error("Timed out")
		}
		readerImpl.msgsToSnd = []*message.Batch{
			message.QuickBatch([][]byte{
				[]byte("foo"),
				[]byte("bar"),
				[]byte("baz"),
				[]byte("buz"),
				[]byte("bev"),
			})}
		select {
		case readerImpl.readChan <- nil:
		case <-time.After(time.Second):
			t.Error("Timed out")
		}
		select {
		case readerImpl.ackChan <- errors.New("ack propagated"):
		case <-time.After(time.Second):
			t.Error("Timed out")
		}
	}()

	require.NoError(t, pres.ConnectWithContext(ctx))

	msg, ackFn, err := pres.ReadWithContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, [][]byte{
		[]byte("foo"),
		[]byte("bar"),
		[]byte("baz"),
		[]byte("buz"),
		[]byte("bev"),
	}, message.GetAllBytes(msg))

	bMsg := message.QuickBatch(nil)
	bMsg.Append(msg.Get(1))
	bMsg.Append(msg.Get(3))
	bMsg.Append(msg.Get(0))
	bMsg.Append(msg.Get(4))
	bMsg.Append(msg.Get(2))

	bErr := batch.NewError(bMsg, errors.New("first"))
	bErr.Failed(1, errors.New("second"))
	bErr.Failed(2, errors.New("third"))

	require.NoError(t, ackFn(ctx, bErr))

	msg, ackFn, err = pres.ReadWithContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, [][]byte{
		[]byte("buz"),
		[]byte("foo"),
	}, message.GetAllBytes(msg))

	require.EqualError(t, ackFn(ctx, nil), "ack propagated")
}

//------------------------------------------------------------------------------

func TestAsyncPreserverBuffer(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	readerImpl := newMockAsyncReaderBlocked()
	pres := input.NewAsyncPreserver(readerImpl)

	sendMsg := func(content string) {
		readerImpl.msgsToSnd = []*message.Batch{message.QuickBatch(
			[][]byte{[]byte(content)},
		)}
		select {
		case readerImpl.readChan <- nil:
		case <-time.After(time.Second):
			t.Error("Timed out")
		}
	}
	sendAck := func() {
		select {
		case readerImpl.ackChan <- nil:
		case <-time.After(time.Second):
			t.Error("Timed out")
		}
	}

	// Send message normally.
	exp := "msg 1"
	exp2 := "msg 2"
	exp3 := "msg 3"

	go sendMsg(exp)
	msg, aFn, err := pres.ReadWithContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if act := string(msg.Get(0).Get()); exp != act {
		t.Errorf("Wrong message returned: %v != %v", act, exp)
	}

	// Prime second message.
	go sendMsg(exp2)

	// Fail previous message, expecting it to be resent.
	_ = aFn(ctx, errors.New("failed"))
	msg, aFn, err = pres.ReadWithContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if act := string(msg.Get(0).Get()); exp != act {
		t.Errorf("Wrong message returned: %v != %v", act, exp)
	}

	// Read the primed message.
	var aFn2 input.AsyncAckFn
	msg, aFn2, err = pres.ReadWithContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if act := string(msg.Get(0).Get()); exp2 != act {
		t.Errorf("Wrong message returned: %v != %v", act, exp2)
	}

	// Fail both messages, expecting them to be resent.
	_ = aFn(ctx, errors.New("failed again"))
	_ = aFn2(ctx, errors.New("failed again"))

	// Read both messages.
	msg, aFn, err = pres.ReadWithContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if act := string(msg.Get(0).Get()); exp != act {
		t.Errorf("Wrong message returned: %v != %v", act, exp)
	}
	msg, aFn2, err = pres.ReadWithContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if act := string(msg.Get(0).Get()); exp2 != act {
		t.Errorf("Wrong message returned: %v != %v", act, exp2)
	}

	// Prime a new message and also an acknowledgement.
	go sendMsg(exp3)
	go sendAck()
	go sendAck()

	// Ack all messages.
	_ = aFn(ctx, nil)
	_ = aFn2(ctx, nil)

	msg, _, err = pres.ReadWithContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if act := string(msg.Get(0).Get()); exp3 != act {
		t.Errorf("Wrong message returned: %v != %v", act, exp3)
	}
}

func TestAsyncPreserverBufferBatchedAcks(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	readerImpl := newMockAsyncReaderBlocked()
	pres := input.NewAsyncPreserver(readerImpl)

	sendMsg := func(content string) {
		readerImpl.msgsToSnd = []*message.Batch{message.QuickBatch(
			[][]byte{[]byte(content)},
		)}
		select {
		case readerImpl.readChan <- nil:
		case <-time.After(time.Second):
			t.Error("Timed out")
		}
	}
	sendAck := func() {
		select {
		case readerImpl.ackChan <- nil:
		case <-time.After(time.Second):
			t.Error("Timed out")
		}
	}

	messages := []string{
		"msg 1",
		"msg 2",
		"msg 3",
	}

	ackFns := []input.AsyncAckFn{}
	for _, exp := range messages {
		go sendMsg(exp)
		msg, aFn, err := pres.ReadWithContext(ctx)
		if err != nil {
			t.Fatal(err)
		}
		ackFns = append(ackFns, aFn)
		if act := string(msg.Get(0).Get()); exp != act {
			t.Errorf("Wrong message returned: %v != %v", act, exp)
		}
	}

	// Fail all messages, expecting them to be resent.
	for _, aFn := range ackFns {
		_ = aFn(ctx, errors.New("failed again"))
	}
	ackFns = []input.AsyncAckFn{}

	for _, exp := range messages {
		msg, aFn, err := pres.ReadWithContext(ctx)
		if err != nil {
			t.Fatal(err)
		}
		ackFns = append(ackFns, aFn)
		if act := string(msg.Get(0).Get()); exp != act {
			t.Errorf("Wrong message returned: %v != %v", act, exp)
		}
	}

	// Ack all messages.
	go func() {
		for _, aFn := range ackFns {
			_ = aFn(ctx, nil)
		}
	}()

	for range ackFns {
		sendAck()
	}
}
