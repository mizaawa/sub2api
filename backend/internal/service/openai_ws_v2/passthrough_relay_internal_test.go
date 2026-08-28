package openai_ws_v2

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	coderws "github.com/coder/websocket"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestRunEntry_DelegatesRelay(t *testing.T) {
	t.Parallel()

	clientConn := newPassthroughTestFrameConn(nil, false)
	upstreamConn := newPassthroughTestFrameConn([]passthroughTestFrame{
		{
			msgType: coderws.MessageText,
			payload: []byte(`{"type":"response.completed","response":{"id":"resp_entry","usage":{"input_tokens":1,"output_tokens":1}}}`),
		},
	}, true)

	result, relayExit := RunEntry(EntryInput{
		Ctx:                context.Background(),
		ClientConn:         clientConn,
		UpstreamConn:       upstreamConn,
		FirstClientMessage: []byte(`{"type":"response.create","model":"gpt-4o","input":[]}`),
	})
	require.Nil(t, relayExit)
	require.Equal(t, "resp_entry", result.RequestID)
}

func TestRunClientToUpstream_ErrorPaths(t *testing.T) {
	t.Parallel()

	t.Run("read client eof", func(t *testing.T) {
		t.Parallel()

		exitCh := make(chan relayExitSignal, 1)
		runClientToUpstream(
			context.Background(),
			newPassthroughTestFrameConn(nil, true),
			nil,
			func(_ coderws.MessageType, _ []byte) error { return nil },
			func() {},
			nil,
			nil,
			exitCh,
		)
		sig := <-exitCh
		require.Equal(t, "read_client", sig.stage)
		require.True(t, sig.graceful)
	})

	t.Run("write upstream failed", func(t *testing.T) {
		t.Parallel()

		exitCh := make(chan relayExitSignal, 1)
		runClientToUpstream(
			context.Background(),
			newPassthroughTestFrameConn([]passthroughTestFrame{
				{msgType: coderws.MessageText, payload: []byte(`{"x":1}`)},
			}, true),
			nil,
			func(_ coderws.MessageType, _ []byte) error { return errors.New("boom") },
			func() {},
			nil,
			nil,
			exitCh,
		)
		sig := <-exitCh
		require.Equal(t, "write_upstream", sig.stage)
		require.False(t, sig.graceful)
	})

	t.Run("forwarded counter and trace callback", func(t *testing.T) {
		t.Parallel()

		exitCh := make(chan relayExitSignal, 1)
		forwarded := &atomic.Int64{}
		traces := make([]RelayTraceEvent, 0, 2)
		runClientToUpstream(
			context.Background(),
			newPassthroughTestFrameConn([]passthroughTestFrame{
				{msgType: coderws.MessageText, payload: []byte(`{"x":1}`)},
			}, true),
			nil,
			func(_ coderws.MessageType, _ []byte) error { return nil },
			func() {},
			forwarded,
			func(event RelayTraceEvent) {
				traces = append(traces, event)
			},
			exitCh,
		)
		sig := <-exitCh
		require.Equal(t, "read_client", sig.stage)
		require.Equal(t, int64(1), forwarded.Load())
		require.NotEmpty(t, traces)
	})
}

func TestRunUpstreamToClient_ErrorAndDropPaths(t *testing.T) {
	t.Parallel()

	t.Run("read upstream eof", func(t *testing.T) {
		t.Parallel()

		exitCh := make(chan relayExitSignal, 1)
		drop := &atomic.Bool{}
		drop.Store(false)
		runUpstreamToClient(
			context.Background(),
			newPassthroughTestFrameConn(nil, true),
			func(_ coderws.MessageType, _ []byte) error { return nil },
			time.Now(),
			time.Now,
			&relayState{},
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			drop,
			nil,
			nil,
			func() {},
			nil,
			exitCh,
		)
		sig := <-exitCh
		require.Equal(t, "read_upstream", sig.stage)
		require.True(t, sig.graceful)
	})

	t.Run("write client failed", func(t *testing.T) {
		t.Parallel()

		exitCh := make(chan relayExitSignal, 1)
		drop := &atomic.Bool{}
		drop.Store(false)
		runUpstreamToClient(
			context.Background(),
			newPassthroughTestFrameConn([]passthroughTestFrame{
				{msgType: coderws.MessageText, payload: []byte(`{"type":"response.output_text.delta","delta":"x"}`)},
			}, true),
			func(_ coderws.MessageType, _ []byte) error { return errors.New("write failed") },
			time.Now(),
			time.Now,
			&relayState{},
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			drop,
			nil,
			nil,
			func() {},
			nil,
			exitCh,
		)
		sig := <-exitCh
		require.Equal(t, "write_client", sig.stage)
	})

	t.Run("drop downstream and stop on terminal", func(t *testing.T) {
		t.Parallel()

		exitCh := make(chan relayExitSignal, 1)
		drop := &atomic.Bool{}
		drop.Store(true)
		dropped := &atomic.Int64{}
		runUpstreamToClient(
			context.Background(),
			newPassthroughTestFrameConn([]passthroughTestFrame{
				{
					msgType: coderws.MessageText,
					payload: []byte(`{"type":"response.completed","response":{"id":"resp_drop","usage":{"input_tokens":1,"output_tokens":1}}}`),
				},
			}, true),
			func(_ coderws.MessageType, _ []byte) error { return nil },
			time.Now(),
			time.Now,
			&relayState{},
			nil,
			nil,
			nil,
			nil,
			nil,
			nil,
			drop,
			nil,
			dropped,
			func() {},
			nil,
			exitCh,
		)
		sig := <-exitCh
		require.Equal(t, "drain_terminal", sig.stage)
		require.True(t, sig.graceful)
		require.Equal(t, int64(1), dropped.Load())
	})
}

func TestRunIdleWatchdog_NoTimeoutWhenDisabled(t *testing.T) {
	t.Parallel()

	exitCh := make(chan relayExitSignal, 1)
	lastActivity := &atomic.Int64{}
	lastActivity.Store(time.Now().UnixNano())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go runIdleWatchdog(ctx, time.Now, 0, lastActivity, nil, exitCh)
	select {
	case <-exitCh:
		t.Fatal("unexpected idle timeout signal")
	case <-time.After(200 * time.Millisecond):
	}
}

func TestHelperFunctionsCoverage(t *testing.T) {
	t.Parallel()

	require.Equal(t, "text", relayMessageTypeString(coderws.MessageText))
	require.Equal(t, "binary", relayMessageTypeString(coderws.MessageBinary))
	require.Contains(t, relayMessageTypeString(coderws.MessageType(99)), "unknown(")

	require.Equal(t, "", relayErrorString(nil))
	require.Equal(t, "x", relayErrorString(errors.New("x")))

	require.True(t, isDisconnectError(io.EOF))
	require.True(t, isDisconnectError(net.ErrClosed))
	require.True(t, isDisconnectError(context.Canceled))
	require.True(t, isDisconnectError(coderws.CloseError{Code: coderws.StatusGoingAway}))
	require.True(t, isDisconnectError(errors.New("broken pipe")))
	require.False(t, isDisconnectError(errors.New("unrelated")))

	require.True(t, isTokenEvent("response.output_text.delta"))
	require.True(t, isTokenEvent("response.output_audio.delta"))
	require.True(t, isTokenEvent("response.completed"))
	require.False(t, isTokenEvent(""))
	require.False(t, isTokenEvent("response.created"))

	require.Equal(t, 2*time.Second, minDuration(2*time.Second, 5*time.Second))
	require.Equal(t, 2*time.Second, minDuration(5*time.Second, 2*time.Second))
	require.Equal(t, 5*time.Second, minDuration(0, 5*time.Second))
	require.Equal(t, 2*time.Second, minDuration(2*time.Second, 0))

	ch := make(chan relayExitSignal, 1)
	ch <- relayExitSignal{stage: "ok"}
	sig, ok := waitRelayExit(ch, 10*time.Millisecond)
	require.True(t, ok)
	require.Equal(t, "ok", sig.stage)
	ch <- relayExitSignal{stage: "ok2"}
	sig, ok = waitRelayExit(ch, 0)
	require.True(t, ok)
	require.Equal(t, "ok2", sig.stage)
	_, ok = waitRelayExit(ch, 10*time.Millisecond)
	require.False(t, ok)

	n, ok := parseUsageIntField(gjson.Get(`{"n":3}`, "n"), true)
	require.True(t, ok)
	require.Equal(t, 3, n)
	_, ok = parseUsageIntField(gjson.Get(`{"n":"x"}`, "n"), true)
	require.False(t, ok)
	n, ok = parseUsageIntField(gjson.Result{}, false)
	require.True(t, ok)
	require.Equal(t, 0, n)
	_, ok = parseUsageIntField(gjson.Result{}, true)
	require.False(t, ok)
}

func TestParseUsageAndEnrichCoverage(t *testing.T) {
	t.Parallel()

	state := &relayState{}
	parseUsageAndAccumulate(state, []byte(`{"type":"response.completed","response":{"usage":{"input_tokens":"bad"}}}`), "response.completed", nil)
	require.Equal(t, 0, state.usage.InputTokens)

	parseUsageAndAccumulate(
		state,
		[]byte(`{"type":"response.completed","response":{"usage":{"input_tokens":9,"output_tokens":"bad","input_tokens_details":{"cached_tokens":2}}}}`),
		"response.completed",
		nil,
	)
	require.Equal(t, 0, state.usage.InputTokens, "部分字段解析失败时不应累加 usage")
	require.Equal(t, 0, state.usage.OutputTokens)
	require.Equal(t, 0, state.usage.CacheReadInputTokens)

	parseUsageAndAccumulate(
		state,
		[]byte(`{"type":"response.completed","response":{"usage":{"input_tokens_details":{"cached_tokens":2}}}}`),
		"response.completed",
		nil,
	)
	require.Equal(t, 0, state.usage.InputTokens, "必填 usage 字段缺失时不应累加 usage")
	require.Equal(t, 0, state.usage.OutputTokens)
	require.Equal(t, 0, state.usage.CacheReadInputTokens)

	parseUsageAndAccumulate(state, []byte(`{"type":"response.completed","response":{"usage":{"input_tokens":2,"output_tokens":1,"input_tokens_details":{"cached_tokens":1,"cache_write_tokens":4},"output_tokens_details":{"image_tokens":3}}}}`), "response.completed", nil)
	finalizeRelayTurnUsage(state)
	require.Equal(t, 2, state.usage.InputTokens)
	require.Equal(t, 1, state.usage.OutputTokens)
	require.Equal(t, 1, state.usage.CacheReadInputTokens)
	require.Equal(t, 4, state.usage.CacheCreationInputTokens)
	require.Equal(t, 3, state.usage.ImageOutputTokens)

	result := &RelayResult{}
	enrichResult(result, state, 5*time.Millisecond)
	require.Equal(t, state.usage.InputTokens, result.Usage.InputTokens)
	require.Equal(t, state.usage.CacheCreationInputTokens, result.Usage.CacheCreationInputTokens)
	require.Equal(t, state.usage.ImageOutputTokens, result.Usage.ImageOutputTokens)
	require.Equal(t, 5*time.Millisecond, result.Duration)
	parseUsageAndAccumulate(state, []byte(`{"type":"response.in_progress","response":{"usage":{"input_tokens":9}}}`), "response.in_progress", nil)
	require.Equal(t, 2, state.usage.InputTokens)
	require.Equal(t, 9, state.turnUsage.InputTokens)
	enrichResult(nil, state, 0)
}

func TestObserveUpstreamMessage_DeduplicatesTerminalUsageByResponseID(t *testing.T) {
	t.Parallel()

	state := &relayState{}
	first := observeUpstreamMessage(state, []byte(`{"type":"response.completed","response":{"id":"resp_dup","usage":{"input_tokens":10,"output_tokens":4}}}`), time.Unix(1, 0), time.Now, nil)
	duplicate := observeUpstreamMessage(state, []byte(`{"type":"response.done","response":{"id":"resp_dup","usage":{"input_tokens":10,"output_tokens":4}}}`), time.Unix(1, 0), time.Now, nil)

	if first.duplicateTerminal {
		t.Fatal("first terminal event must not be marked duplicate")
	}
	if !duplicate.terminal || !duplicate.duplicateTerminal {
		t.Fatalf("duplicate terminal event should still drain but not bill: %+v", duplicate)
	}
	if state.usage.InputTokens != 10 || state.usage.OutputTokens != 4 {
		t.Fatalf("duplicate terminal usage was accumulated: %+v", state.usage)
	}
	turns := 0
	emitTurnComplete(func(RelayTurnResult) { turns++ }, state, first)
	emitTurnComplete(func(RelayTurnResult) { turns++ }, state, duplicate)
	if turns != 1 {
		t.Fatalf("expected one turn callback, got %d", turns)
	}
}

func TestObserveUpstreamMessage_DeduplicatesTerminalUsageWhenResponseIDChanges(t *testing.T) {
	t.Parallel()

	state := &relayState{}
	created := observeUpstreamMessage(
		state,
		[]byte(`{"type":"response.created","response":{"id":"resp_original","model":"gpt-5.3"}}`),
		time.Unix(1, 0),
		time.Now,
		nil,
	)
	if created.terminal {
		t.Fatal("response.created must not be terminal")
	}
	first := observeUpstreamMessage(
		state,
		[]byte(`{"type":"response.completed","response":{"id":"resp_original","usage":{"input_tokens":10,"output_tokens":4}}}`),
		time.Unix(1, 0),
		time.Now,
		nil,
	)
	mutated := observeUpstreamMessage(
		state,
		[]byte(`{"type":"response.completed","response":{"id":"resp_mutated","usage":{"input_tokens":999999,"output_tokens":888888}}}`),
		time.Unix(1, 0),
		time.Now,
		nil,
	)

	require.True(t, first.terminal)
	require.False(t, first.duplicateTerminal)
	require.True(t, mutated.terminal)
	require.True(t, mutated.duplicateTerminal)
	require.Equal(t, Usage{InputTokens: 10, OutputTokens: 4}, state.usage)
}

func TestObserveUpstreamMessage_DeduplicatesAnonymousTerminalUsage(t *testing.T) {
	t.Parallel()

	state := &relayState{}
	first := observeUpstreamMessage(
		state,
		[]byte(`{"type":"response.completed","response":{"usage":{"input_tokens":10,"output_tokens":4}}}`),
		time.Unix(1, 0),
		time.Now,
		nil,
	)
	duplicate := observeUpstreamMessage(
		state,
		[]byte(`{"type":"response.done","response":{"usage":{"input_tokens":999999,"output_tokens":888888}}}`),
		time.Unix(1, 0),
		time.Now,
		nil,
	)

	require.True(t, first.terminal)
	require.Empty(t, first.responseID)
	require.False(t, first.duplicateTerminal)
	require.True(t, duplicate.terminal)
	require.True(t, duplicate.duplicateTerminal)
	require.Equal(t, Usage{InputTokens: 10, OutputTokens: 4}, state.usage)
}

func TestRelayState_ResetAnonymousTerminalStartsNextTurn(t *testing.T) {
	t.Parallel()

	state := &relayState{}
	first := observeUpstreamMessage(
		state,
		[]byte(`{"type":"response.completed","response":{"usage":{"input_tokens":10,"output_tokens":4}}}`),
		time.Unix(1, 0),
		time.Now,
		nil,
	)
	require.True(t, first.terminal)
	require.True(t, state.anonymousTerminalSeen.Load())

	state.resetAnonymousTerminal()
	next := observeUpstreamMessage(
		state,
		[]byte(`{"type":"response.completed","response":{"usage":{"input_tokens":3,"output_tokens":2}}}`),
		time.Unix(1, 0),
		time.Now,
		nil,
	)
	require.True(t, next.terminal)
	require.False(t, next.duplicateTerminal)
	require.Equal(t, Usage{InputTokens: 13, OutputTokens: 6}, state.usage)
}

func TestObserveUpstreamMessage_DeduplicatesAnonymousAfterIDTerminal(t *testing.T) {
	t.Parallel()

	state := &relayState{}
	withID := observeUpstreamMessage(
		state,
		[]byte(`{"type":"response.completed","response":{"id":"resp_with_id","usage":{"input_tokens":10,"output_tokens":4}}}`),
		time.Unix(1, 0),
		time.Now,
		nil,
	)
	require.True(t, withID.terminal)
	require.False(t, state.anonymousTerminalSeen.Load())

	anonymous := observeUpstreamMessage(
		state,
		[]byte(`{"type":"response.done","response":{"usage":{"input_tokens":3,"output_tokens":2}}}`),
		time.Unix(1, 0),
		time.Now,
		nil,
	)
	require.True(t, anonymous.terminal)
	require.True(t, anonymous.duplicateTerminal)
	require.Equal(t, Usage{InputTokens: 10, OutputTokens: 4}, state.usage)
}

func TestParseUsageAndAccumulateRejectsAggregateOverflow(t *testing.T) {
	t.Parallel()
	state := &relayState{usage: Usage{InputTokens: maxReportedUsageTokens - 1}}
	got := parseUsageAndAccumulate(state, []byte(`{"type":"response.done","response":{"usage":{"input_tokens":2,"output_tokens":1}}}`), "response.done", nil)
	require.Equal(t, Usage{InputTokens: 2, OutputTokens: 1}, got)
	require.Equal(t, Usage{}, finalizeRelayTurnUsage(state))
	if state.usage.InputTokens != maxReportedUsageTokens-1 || state.usage.OutputTokens != 0 {
		t.Fatalf("aggregate usage changed after overflow: %+v", state.usage)
	}
}

func TestParseUsageAndAccumulateAcceptsChatUsageAliases(t *testing.T) {
	t.Parallel()

	state := &relayState{}
	got := parseUsageAndAccumulate(
		state,
		[]byte(`{"type":"response.done","response":{"usage":{"prompt_tokens":12,"completion_tokens":6,"prompt_tokens_details":{"cached_tokens":4},"completion_tokens_details":{"image_tokens":2}}}}`),
		"response.done",
		nil,
	)
	require.Equal(t, 12, got.InputTokens)
	require.Equal(t, 6, got.OutputTokens)
	require.Equal(t, 4, got.CacheReadInputTokens)
	require.Equal(t, 2, got.ImageOutputTokens)
	require.Equal(t, got, finalizeRelayTurnUsage(state))
	require.Equal(t, got, state.usage)
}

func TestParseUsageAndAccumulateRejectsFractionalAndMalformedNestedFields(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		usage string
	}{
		{name: "fractional input", usage: `{"input_tokens":1.9,"output_tokens":2}`},
		{name: "fractional image", usage: `{"input_tokens":1,"output_tokens":2,"output_tokens_details":{"image_tokens":3.5}}`},
		{name: "string cache", usage: `{"input_tokens":1,"output_tokens":2,"input_tokens_details":{"cached_tokens":"4"}}`},
		{name: "huge exponent cache", usage: `{"input_tokens":1,"output_tokens":2,"input_tokens_details":{"cache_write_tokens":1e999}}`},
		{name: "negative nested", usage: `{"input_tokens":1,"output_tokens":2,"prompt_tokens_details":{"cached_tokens":-1}}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state := &relayState{}
			got := parseUsageAndAccumulate(state, []byte(`{"type":"response.done","response":{"usage":`+tc.usage+`}}`), "response.done", nil)
			require.Equal(t, Usage{}, got)
			require.Equal(t, Usage{}, state.usage)
		})
	}
}

func TestObserveUpstreamMessage_DoesNotReopenTerminalAfterInterveningData(t *testing.T) {
	t.Parallel()

	state := &relayState{}
	first := observeUpstreamMessage(state,
		[]byte(`{"type":"response.completed","response":{"id":"resp_a","usage":{"input_tokens":10,"output_tokens":4}}}`),
		time.Unix(1, 0), time.Now, nil)
	require.True(t, first.terminal)
	require.False(t, first.duplicateTerminal)

	// A forged data/lifecycle event must not reset the one-terminal guard.
	nonTerminal := observeUpstreamMessage(state,
		[]byte(`{"type":"response.output_text.delta","response":{"id":"resp_a"},"delta":"x"}`),
		time.Unix(1, 0), time.Now, nil)
	require.False(t, nonTerminal.terminal)

	mutated := observeUpstreamMessage(state,
		[]byte(`{"type":"response.done","response":{"id":"resp_b","usage":{"input_tokens":999,"output_tokens":888}}}`),
		time.Unix(1, 0), time.Now, nil)
	require.True(t, mutated.terminal)
	require.True(t, mutated.duplicateTerminal)
	require.Equal(t, Usage{InputTokens: 10, OutputTokens: 4}, state.usage)
}

func TestOpenAICacheCreationTokensFromUsageNestedZeroWins(t *testing.T) {
	t.Parallel()

	usage := gjson.Parse(`{"input_tokens_details":{"cache_write_tokens":0},"cache_creation_input_tokens":19}`)
	got, ok := openAICacheCreationTokensFromUsage(usage)
	require.True(t, ok)
	require.Zero(t, got)
}

func TestEmitTurnCompleteCoverage(t *testing.T) {
	t.Parallel()

	// 非 terminal 事件不应触发。
	called := 0
	emitTurnComplete(func(turn RelayTurnResult) {
		called++
	}, &relayState{requestModel: "gpt-5"}, observedUpstreamEvent{
		terminal:   false,
		eventType:  "response.output_text.delta",
		responseID: "resp_ignored",
		usage:      Usage{InputTokens: 1},
	})
	require.Equal(t, 0, called)

	// 缺少 response_id 时不应触发。
	emitTurnComplete(func(turn RelayTurnResult) {
		called++
	}, &relayState{requestModel: "gpt-5"}, observedUpstreamEvent{
		terminal:  true,
		eventType: "response.completed",
	})
	require.Equal(t, 0, called)

	// terminal 且 response_id 存在，应该触发；state=nil 时 model 为空串。
	var got RelayTurnResult
	emitTurnComplete(func(turn RelayTurnResult) {
		called++
		got = turn
	}, nil, observedUpstreamEvent{
		terminal:   true,
		eventType:  "response.completed",
		responseID: "resp_emit",
		usage:      Usage{InputTokens: 2, OutputTokens: 3},
	})
	require.Equal(t, 1, called)
	require.Equal(t, "resp_emit", got.RequestID)
	require.Equal(t, "response.completed", got.TerminalEventType)
	require.Equal(t, 2, got.Usage.InputTokens)
	require.Equal(t, 3, got.Usage.OutputTokens)
	require.Equal(t, "", got.RequestModel)
}

func TestIsDisconnectErrorCoverage_CloseStatusesAndMessageBranches(t *testing.T) {
	t.Parallel()

	require.True(t, isDisconnectError(coderws.CloseError{Code: coderws.StatusNormalClosure}))
	require.True(t, isDisconnectError(coderws.CloseError{Code: coderws.StatusNoStatusRcvd}))
	require.True(t, isDisconnectError(coderws.CloseError{Code: coderws.StatusAbnormalClosure}))
	require.True(t, isDisconnectError(errors.New("connection reset by peer")))
	require.False(t, isDisconnectError(errors.New("   ")))
}

func TestIsTokenEventCoverageBranches(t *testing.T) {
	t.Parallel()

	require.False(t, isTokenEvent("response.in_progress"))
	require.False(t, isTokenEvent("response.output_item.added"))
	require.True(t, isTokenEvent("response.output_audio.delta"))
	require.True(t, isTokenEvent("response.output"))
	require.True(t, isTokenEvent("response.done"))
}

func TestShouldParseUsageTerminalEvents(t *testing.T) {
	t.Parallel()

	for _, eventType := range []string{
		"response.completed",
		"response.done",
		"response.failed",
		"response.incomplete",
		"response.cancelled",
		"response.canceled",
	} {
		require.True(t, shouldParseUsage(eventType), eventType)
	}
	require.False(t, shouldParseUsage("response.output_text.delta"))
	require.False(t, shouldParseUsage(""))
}

func TestRelayTurnTimingHelpersCoverage(t *testing.T) {
	t.Parallel()

	now := time.Unix(100, 0)
	// nil state
	require.Nil(t, openAIWSRelayGetOrInitTurnTiming(nil, "resp_nil", now))
	_, ok := openAIWSRelayDeleteTurnTiming(nil, "resp_nil")
	require.False(t, ok)

	state := &relayState{}
	timing := openAIWSRelayGetOrInitTurnTiming(state, "resp_a", now)
	require.NotNil(t, timing)
	require.Equal(t, now, timing.startAt)

	// 再次获取返回同一条 timing
	timing2 := openAIWSRelayGetOrInitTurnTiming(state, "resp_a", now.Add(5*time.Second))
	require.NotNil(t, timing2)
	require.Equal(t, now, timing2.startAt)

	// 删除存在键
	deleted, ok := openAIWSRelayDeleteTurnTiming(state, "resp_a")
	require.True(t, ok)
	require.Equal(t, now, deleted.startAt)

	// 删除不存在键
	_, ok = openAIWSRelayDeleteTurnTiming(state, "resp_a")
	require.False(t, ok)
}

func TestObserveUpstreamMessage_ResponseModelIsTurnLocalAndTerminalWins(t *testing.T) {
	t.Parallel()

	state := &relayState{requestModel: "gpt-5.6-sol"}
	startAt := time.Unix(0, 0)
	now := startAt
	nowFn := func() time.Time {
		now = now.Add(5 * time.Millisecond)
		return now
	}

	created := observeUpstreamMessage(
		state,
		[]byte(`{"type":"response.created","response":{"id":"resp_1","model":"gpt-5.5"}}`),
		startAt,
		nowFn,
		nil,
	)
	require.False(t, created.terminal)

	completed := observeUpstreamMessage(
		state,
		[]byte(`{"type":"response.completed","response":{"id":"resp_1","model":"gpt-5.4","usage":{"input_tokens":1,"output_tokens":2}}}`),
		startAt,
		nowFn,
		nil,
	)
	require.True(t, completed.terminal)
	require.Equal(t, "gpt-5.4", completed.responseModel)
	require.True(t, completed.responseConflict)

	var firstTurn RelayTurnResult
	emitTurnComplete(func(turn RelayTurnResult) { firstTurn = turn }, state, completed)
	require.Equal(t, "gpt-5.4", firstTurn.ResponseModel)
	require.True(t, firstTurn.ResponseModelConflict)

	state.startClientTurn()
	observeUpstreamMessage(
		state,
		[]byte(`{"type":"response.created","response":{"id":"resp_2","model":"gpt-5.3"}}`),
		startAt,
		nowFn,
		nil,
	)
	second := observeUpstreamMessage(
		state,
		[]byte(`{"type":"response.completed","response":{"id":"resp_2","model":"GPT-5.3","usage":{"input_tokens":3,"output_tokens":4}}}`),
		startAt,
		nowFn,
		nil,
	)
	require.Equal(t, "GPT-5.3", second.responseModel)
	require.False(t, second.responseConflict, "the previous turn must not contaminate this turn")
}

func TestObserveUpstreamMessage_ForgedLifecycleCannotReopenTerminal(t *testing.T) {
	t.Parallel()

	state := &relayState{}
	first := observeUpstreamMessage(
		state,
		[]byte(`{"type":"response.completed","response":{"id":"resp_real","usage":{"input_tokens":10,"output_tokens":4}}}`),
		time.Unix(1, 0),
		time.Now,
		nil,
	)
	require.True(t, first.terminal)

	// The upstream lifecycle frame is metadata only. It must not act as a
	// second turn boundary after the client has already received a terminal.
	created := observeUpstreamMessage(
		state,
		[]byte(`{"type":"response.created","response":{"id":"resp_forged"}}`),
		time.Unix(1, 0),
		time.Now,
		nil,
	)
	require.False(t, created.terminal)

	duplicate := observeUpstreamMessage(
		state,
		[]byte(`{"type":"response.completed","response":{"id":"resp_forged","usage":{"input_tokens":999,"output_tokens":888}}}`),
		time.Unix(1, 0),
		time.Now,
		nil,
	)
	require.True(t, duplicate.terminal)
	require.True(t, duplicate.duplicateTerminal)
	require.Equal(t, Usage{InputTokens: 10, OutputTokens: 4}, state.usage)
}

func TestObserveUpstreamMessage_ResponseIDFallbackPolicy(t *testing.T) {
	t.Parallel()

	state := &relayState{requestModel: "gpt-5"}
	startAt := time.Unix(0, 0)
	now := startAt
	nowFn := func() time.Time {
		now = now.Add(5 * time.Millisecond)
		return now
	}

	// 非 terminal：仅有顶层 id，不应把 event id 当成 response_id。
	observed := observeUpstreamMessage(
		state,
		[]byte(`{"type":"response.output_text.delta","id":"evt_123","delta":"hi"}`),
		startAt,
		nowFn,
		nil,
	)
	require.False(t, observed.terminal)
	require.Equal(t, "", observed.responseID)

	// terminal：允许兜底用顶层 id（用于兼容少数字段变体）。
	observed = observeUpstreamMessage(
		state,
		[]byte(`{"type":"response.completed","id":"resp_fallback","response":{"usage":{"input_tokens":1,"output_tokens":1}}}`),
		startAt,
		nowFn,
		nil,
	)
	require.True(t, observed.terminal)
	require.Equal(t, "resp_fallback", observed.responseID)
}

func TestRelayUsage_NonTerminalSnapshotsAreNotAdded(t *testing.T) {
	t.Parallel()

	state := &relayState{}
	start := time.Unix(1, 0)
	now := func() time.Time { return start }

	first := observeUpstreamMessage(state, []byte(`{"type":"response.in_progress","response":{"id":"resp_snapshot","usage":{"input_tokens":10,"output_tokens":2}}}`), start, now, nil)
	second := observeUpstreamMessage(state, []byte(`{"type":"response.output_item.done","response":{"id":"resp_snapshot","usage":{"input_tokens":15,"output_tokens":3}}}`), start, now, nil)
	require.False(t, first.terminal)
	require.False(t, second.terminal)
	require.Equal(t, Usage{}, state.usage)
	require.Equal(t, Usage{InputTokens: 15, OutputTokens: 3}, state.turnUsage)

	terminal := observeUpstreamMessage(state, []byte(`{"type":"response.completed","response":{"id":"resp_snapshot"}}`), start, now, nil)
	require.True(t, terminal.terminal)
	require.Equal(t, Usage{InputTokens: 15, OutputTokens: 3}, terminal.usage)
	require.Equal(t, terminal.usage, state.usage)
	require.Equal(t, Usage{}, state.turnUsage)
}

func TestRelayState_StartClientTurnDropsPreviousTurnUsage(t *testing.T) {
	t.Parallel()

	state := &relayState{
		turnUsage: Usage{InputTokens: 15, OutputTokens: 3},
		pendingBareError: &observedUpstreamEvent{
			eventType: "error",
			usage:     Usage{InputTokens: 7, OutputTokens: 2},
		},
	}

	state.startClientTurn()

	require.Equal(t, Usage{}, state.turnUsage)
	require.Nil(t, state.pendingBareError)

	// A terminal snapshot in the new turn must be billed on its own.
	next := observeUpstreamMessage(
		state,
		[]byte(`{"type":"response.completed","response":{"id":"resp_next","usage":{"input_tokens":4,"output_tokens":1}}}`),
		time.Unix(1, 0),
		time.Now,
		nil,
	)
	require.True(t, next.terminal)
	require.Equal(t, Usage{InputTokens: 4, OutputTokens: 1}, next.usage)
	require.Equal(t, Usage{InputTokens: 4, OutputTokens: 1}, state.usage)
}

func TestRelayUsage_EmptyTerminalKeepsPriorSnapshot(t *testing.T) {
	t.Parallel()

	state := &relayState{}
	start := time.Unix(2, 0)
	now := func() time.Time { return start }
	observeUpstreamMessage(state, []byte(`{"type":"response.in_progress","response":{"id":"resp_empty","usage":{"input_tokens":9,"output_tokens":4,"input_tokens_details":{"cached_tokens":2}}}}`), start, now, nil)

	terminal := observeUpstreamMessage(state, []byte(`{"type":"response.completed","response":{"id":"resp_empty"}}`), start, now, nil)
	require.True(t, terminal.terminal)
	require.Equal(t, Usage{InputTokens: 9, OutputTokens: 4, CacheReadInputTokens: 2}, terminal.usage)
	require.Equal(t, terminal.usage, state.usage)
}

func TestRelayUsage_ErrorThenFailedSettlesOnce(t *testing.T) {
	t.Parallel()

	state := &relayState{}
	start := time.Unix(3, 0)
	now := func() time.Time { return start }
	observeUpstreamMessage(state, []byte(`{"type":"response.in_progress","response":{"id":"resp_failed","usage":{"input_tokens":11,"output_tokens":5}}}`), start, now, nil)

	bareError := observeUpstreamMessage(state, []byte(`{"type":"error","error":{"message":"upstream"},"usage":{"input_tokens":7,"output_tokens":2}}`), start, now, nil)
	require.False(t, bareError.terminal)
	require.NotNil(t, state.pendingBareError)

	failed := observeUpstreamMessage(state, []byte(`{"type":"response.failed","response":{"id":"resp_failed","usage":{"input_tokens":8,"output_tokens":3}}}`), start, now, nil)
	require.True(t, failed.terminal)
	require.Equal(t, Usage{InputTokens: 8, OutputTokens: 3}, failed.usage)
	require.Nil(t, state.pendingBareError)

	turns := 0
	emitTurnComplete(func(RelayTurnResult) { turns++ }, state, bareError)
	emitTurnComplete(func(RelayTurnResult) { turns++ }, state, failed)
	require.Equal(t, 1, turns)
	require.Equal(t, Usage{InputTokens: 8, OutputTokens: 3}, state.usage)
}

func TestRelayState_ConcurrentTransitionsAreSerialized(t *testing.T) {
	state := &relayState{requestModel: "gpt-4o"}
	const iterations = 200
	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			state.startClientTurn()
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			observeUpstreamMessage(
				state,
				[]byte(`{"type":"response.in_progress","response":{"id":"resp_concurrent","usage":{"input_tokens":3}}}`),
				time.Unix(1, 0),
				time.Now,
				nil,
			)
			observeUpstreamMessage(
				state,
				[]byte(`{"type":"response.completed","response":{"id":"resp_concurrent","usage":{"input_tokens":3,"output_tokens":2}}}`),
				time.Unix(1, 0),
				time.Now,
				nil,
			)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			finalizePendingBareError(state, time.Now())
			finalizeRelayTurnUsage(state)
		}
	}()
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			var result RelayResult
			enrichResult(&result, state, time.Millisecond)
		}
	}()

	wg.Wait()
	var result RelayResult
	enrichResult(&result, state, 0)
	for _, value := range []int{
		result.Usage.InputTokens,
		result.Usage.OutputTokens,
		result.Usage.CacheCreationInputTokens,
		result.Usage.CacheReadInputTokens,
		result.Usage.ImageOutputTokens,
	} {
		require.GreaterOrEqual(t, value, 0)
		require.LessOrEqual(t, value, maxReportedUsageTokens)
	}
}
