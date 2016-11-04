package transport

import (
	"errors"
	"logging"
	"time"
)

/*
 * Constants
 */
var (
	TCP_STATE_DEFAULT_TIMEOUT_MILLIS = 250
	TCP_MSL_MILLIS                   = 60 * 1000
	TCP_MAX_RETRY_COUNT              = 3

	// EVENTS
	TCP_ACTIVE_OPEN  TcpTransitionEvent = TcpTransitionEvent{ActiveOpen: true}
	TCP_PASSIVE_OPEN TcpTransitionEvent = TcpTransitionEvent{PassiveOpen: true}
	TCP_SEND         TcpTransitionEvent = TcpTransitionEvent{Send: true}
	TCP_CLOSE        TcpTransitionEvent = TcpTransitionEvent{Close: true}
	TCP_RECV_SYN     TcpTransitionEvent = TcpTransitionEvent{RecvSyn: true}
	TCP_RECV_ACK     TcpTransitionEvent = TcpTransitionEvent{RecvAck: true}
	TCP_RECV_FIN     TcpTransitionEvent = TcpTransitionEvent{RecvFin: true}
	TCP_RECV_SYN_ACK TcpTransitionEvent = TcpTransitionEvent{RecvSyn: true, RecvAck: true}
	TCP_TIMEOUT_2MSL TcpTransitionEvent = TcpTransitionEvent{Timeout: true}

	// RESPONSES
	TCP_RESP_DO_NOTHING   TcpTransitionResponse = TcpTransitionResponse{}
	TCP_RESP_SEND_SYN     TcpTransitionResponse = TcpTransitionResponse{ShouldSendSyn: true}
	TCP_RESP_SEND_ACK     TcpTransitionResponse = TcpTransitionResponse{ShouldSendAck: true}
	TCP_RESP_SEND_FIN     TcpTransitionResponse = TcpTransitionResponse{ShouldSendFin: true}
	TCP_RESP_SEND_SYN_ACK TcpTransitionResponse = TcpTransitionResponse{ShouldSendSyn: true, ShouldSendAck: true}
	TCP_RESP_DEL_SOCK     TcpTransitionResponse = TcpTransitionResponse{ShouldDeleteSocket: true}

	// STATES
	// establish connection
	TCP_INITIAL_CLOSED TcpState = TcpState{Name: "CLOSED", StateTimeoutMillis: 0}
	TCP_LISTEN         TcpState = TcpState{Name: "LISTEN", StateTimeoutMillis: 0}
	TCP_SYN_RCVD       TcpState = TcpState{Name: "SYN_RCVD", StateTimeoutMillis: TCP_STATE_DEFAULT_TIMEOUT_MILLIS}
	TCP_SYN_SENT       TcpState = TcpState{Name: "SYN_SENT", StateTimeoutMillis: TCP_STATE_DEFAULT_TIMEOUT_MILLIS}
	TCP_ESTAB          TcpState = TcpState{Name: "ESTAB", StateTimeoutMillis: 0}

	// active close
	TCP_FIN_WAIT_1 TcpState = TcpState{Name: "FIN_WAIT_1", IsActiveClose: true, StateTimeoutMillis: TCP_STATE_DEFAULT_TIMEOUT_MILLIS}
	TCP_FIN_WAIT_2 TcpState = TcpState{Name: "FIN_WAIT_2", IsActiveClose: true, StateTimeoutMillis: TCP_STATE_DEFAULT_TIMEOUT_MILLIS}
	TCP_CLOSING    TcpState = TcpState{Name: "CLOSING", IsActiveClose: true, StateTimeoutMillis: TCP_STATE_DEFAULT_TIMEOUT_MILLIS}
	TCP_TIME_WAIT  TcpState = TcpState{Name: "TIME_WAIT", IsActiveClose: true, StateTimeoutMillis: TCP_MSL_MILLIS * 2}

	// passive close
	TCP_CLOSE_WAIT   TcpState = TcpState{Name: "CLOSE_WAIT", IsPassiveClose: true, StateTimeoutMillis: 0}
	TCP_LAST_ACK     TcpState = TcpState{Name: "LAST_ACK", IsPassiveClose: true, StateTimeoutMillis: TCP_STATE_DEFAULT_TIMEOUT_MILLIS}
	TCP_FINAL_CLOSED TcpState = TcpState{Name: "CLOSED", IsPassiveClose: true, StateTimeoutMillis: 0}
)

type TcpTransitionEvent struct {
	ActiveOpen  bool
	PassiveOpen bool
	Send        bool
	Close       bool
	RecvSyn     bool
	RecvAck     bool
	RecvFin     bool
	Timeout     bool
}

func MakeTcpTransitionEvent(header TCPHeader) TcpTransitionEvent {
	recvSyn, recvAck, recvFin := false, false, false
	if header.HasFlag(SYN) {
		recvSyn = true
	}

	if header.HasFlag(ACK) {
		recvAck = true
	}

	if header.HasFlag(FIN) {
		recvFin = true
	}

	return TcpTransitionEvent{RecvSyn: recvSyn, RecvAck: recvAck, RecvFin: recvFin}
}

type TcpTransitionResponse struct {
	ShouldSendSyn      bool
	ShouldSendAck      bool
	ShouldSendFin      bool
	ShouldDeleteSocket bool
}

func (r *TcpTransitionResponse) ShouldDoNothing() bool {
	return !(r.ShouldSendSyn || r.ShouldSendAck || r.ShouldSendFin || r.ShouldDeleteSocket)
}

func (r *TcpTransitionResponse) GetCtrlFlags() int {
	flags := 0
	if r.ShouldSendSyn {
		flags = flags | SYN
	}

	if r.ShouldSendAck {
		flags = flags | ACK
	}

	if r.ShouldSendFin {
		flags = flags | FIN
	}

	return flags
}

type TcpState struct {
	Name               string
	IsActiveClose      bool
	IsPassiveClose     bool
	StateTimeoutMillis int
	CloseOnTimeout     bool
}

func (state TcpState) CanTimeout() bool {
	return state.StateTimeoutMillis > 0
}

type TcpTransition struct {
	CurrentState TcpState
	Event        TcpTransitionEvent
}

type TcpStateMachine struct {
	fd               int
	currentState     TcpState
	previousResponse TcpTransitionResponse
	states           map[TcpTransition]TcpState
	responses        map[TcpTransition]TcpTransitionResponse

	stateTimer *time.Timer
	retryCount int
}

func MakeTcpStateMachine(fd int, initialState TcpState, states map[TcpTransition]TcpState, responses map[TcpTransition]TcpTransitionResponse) TcpStateMachine {
	emptyTimer := time.NewTimer(time.Duration(TCP_STATE_DEFAULT_TIMEOUT_MILLIS) * time.Millisecond)
	if !emptyTimer.Stop() {
		emptyTimer = time.NewTimer(time.Duration(1 * time.Hour))
	}

	return TcpStateMachine{
		fd:               fd,
		currentState:     initialState,
		previousResponse: TCP_RESP_DO_NOTHING,
		states:           states,
		responses:        responses,
		stateTimer:       emptyTimer,
	}
}

func (m *TcpStateMachine) CurrentState() TcpState {
	return m.currentState
}

func (m *TcpStateMachine) RetryCount() int {
	return m.retryCount
}

func (m *TcpStateMachine) IncrementRetryCount() {
	m.retryCount += 1
}

func (m *TcpStateMachine) TimerChannel() <-chan time.Time {
	return m.stateTimer.C
}

func (m *TcpStateMachine) ResetStateTimer() {
	m.stateTimer = time.NewTimer(time.Duration(m.CurrentState().StateTimeoutMillis) * time.Millisecond)
}

func (m *TcpStateMachine) HasTransition(event TcpTransitionEvent) bool {
	transition := TcpTransition{m.CurrentState(), event}
	if _, ok := m.states[transition]; ok {
		return true
	} else {
		return false
	}
}

func (m *TcpStateMachine) GetResponse(event TcpTransitionEvent) (TcpTransitionResponse, error) {
	var response TcpTransitionResponse
	if !m.HasTransition(event) {
		return response, errors.New("Statemachine does not have transition for the given input event")
	}

	transition := TcpTransition{m.CurrentState(), event}
	return m.responses[transition], nil
}

func (m *TcpStateMachine) GetPreviousResponse() TcpTransitionResponse {
	return m.previousResponse
}

func (m *TcpStateMachine) Transit(event TcpTransitionEvent) error {
	if !m.HasTransition(event) {
		return errors.New("Statemachine does not have transition for the given input event")
	}

	transition := TcpTransition{m.CurrentState(), event}
	m.currentState = m.states[transition]
	m.previousResponse = m.responses[transition]
	m.retryCount = 0

	if m.CurrentState().CanTimeout() {
		m.stateTimer = time.NewTimer(time.Duration(m.CurrentState().StateTimeoutMillis) * time.Millisecond)
	}

	logging.Logger.Println("[TcpStateMachine]", m.fd, "has transited into state", m.CurrentState().Name)
	return nil
}

type TcpStateMachineBuilder struct {
	initialState TcpState
	states       map[TcpTransition]TcpState
	responses    map[TcpTransition]TcpTransitionResponse
}

func MakeTcpStateMachineBuilder(initialState TcpState) TcpStateMachineBuilder {
	states := make(map[TcpTransition]TcpState)
	responses := make(map[TcpTransition]TcpTransitionResponse)

	return TcpStateMachineBuilder{initialState, states, responses}
}

func (b *TcpStateMachineBuilder) RegisterTransition(fromState TcpState, event TcpTransitionEvent, response TcpTransitionResponse, toState TcpState) {
	transition := TcpTransition{fromState, event}
	b.states[transition] = toState
	b.responses[transition] = response
}

func (b *TcpStateMachineBuilder) Build(fd int) TcpStateMachine {
	return MakeTcpStateMachine(fd, b.initialState, b.states, b.responses)
}
