package transport

import (
	"fmt"
)

func TestStateMachine() {
	builder := MakeTcpStateMachineBuilder(TCP_INITIAL_CLOSED)

	builder.RegisterTransition(TCP_INITIAL_CLOSED, TCP_PASSIVE_OPEN, TCP_RESP_DO_NOTHING, TCP_LISTEN)

	builder.RegisterTransition(TCP_LISTEN, TCP_CLOSE, TCP_RESP_DEL_SOCK, TCP_INITIAL_CLOSED)
	builder.RegisterTransition(TCP_LISTEN, TCP_RECV_SYN, TCP_RESP_SEND_SYN_ACK, TCP_SYN_RCVD)
	builder.RegisterTransition(TCP_LISTEN, TCP_SEND, TCP_RESP_SEND_SYN, TCP_SYN_SENT)

	builder.RegisterTransition(TCP_SYN_RCVD, TCP_CLOSE, TCP_RESP_SEND_FIN, TCP_FIN_WAIT_1)
	builder.RegisterTransition(TCP_SYN_RCVD, TCP_RECV_ACK, TCP_RESP_DO_NOTHING, TCP_ESTAB)

	builder.RegisterTransition(TCP_SYN_SENT, TCP_CLOSE, TCP_RESP_DEL_SOCK, TCP_INITIAL_CLOSED)
	builder.RegisterTransition(TCP_SYN_SENT, TCP_RECV_SYN, TCP_RESP_SEND_SYN_ACK, TCP_SYN_RCVD)
	builder.RegisterTransition(TCP_SYN_SENT, TCP_RECV_SYN_ACK, TCP_RESP_SEND_ACK, TCP_ESTAB)

	builder.RegisterTransition(TCP_ESTAB, TCP_CLOSE, TCP_RESP_SEND_FIN, TCP_FIN_WAIT_1)
	builder.RegisterTransition(TCP_ESTAB, TCP_RECV_FIN, TCP_RESP_SEND_ACK, TCP_CLOSE_WAIT)

	builder.RegisterTransition(TCP_FIN_WAIT_1, TCP_RECV_ACK, TCP_RESP_DO_NOTHING, TCP_FIN_WAIT_2)
	builder.RegisterTransition(TCP_FIN_WAIT_1, TCP_RECV_FIN, TCP_RESP_SEND_ACK, TCP_CLOSING)

	builder.RegisterTransition(TCP_FIN_WAIT_2, TCP_RECV_FIN, TCP_RESP_SEND_ACK, TCP_TIME_WAIT)

	builder.RegisterTransition(TCP_CLOSING, TCP_RECV_ACK, TCP_RESP_DO_NOTHING, TCP_TIME_WAIT)

	builder.RegisterTransition(TCP_TIME_WAIT, TCP_TIMEOUT_2MSL, TCP_RESP_DEL_SOCK, TCP_FINAL_CLOSED)

	builder.RegisterTransition(TCP_CLOSE_WAIT, TCP_CLOSE, TCP_RESP_SEND_FIN, TCP_LAST_ACK)

	builder.RegisterTransition(TCP_LAST_ACK, TCP_RECV_ACK, TCP_RESP_DO_NOTHING, TCP_FINAL_CLOSED)

	machine := builder.Build(1)

	machine.CurrentState()
	var r TcpTransitionResponse

	fmt.Printf("Starting state is %s\n", machine.CurrentState().Name)
	fmt.Printf("\n\n")

	fmt.Printf("Passive OPEN, expecte next state to be LISTEN\n")
	machine.Transit(TCP_PASSIVE_OPEN)
	fmt.Printf("Current state is %s\n", machine.CurrentState().Name)
	fmt.Printf("\n\n")

	fmt.Printf("CLOSE, expect next state to be CLOSED\n")
	machine.Transit(TCP_CLOSE)
	fmt.Printf("Current state is %s\n", machine.CurrentState().Name)
	fmt.Printf("\n\n")

	fmt.Printf("Passive OPEN, expecte next state to be LISTEN\n")
	machine.Transit(TCP_PASSIVE_OPEN)
	fmt.Printf("Current state is %s\n", machine.CurrentState().Name)
	fmt.Printf("\n\n")

	fmt.Printf("SEND, expect next state to be SYN_SENT\n")
	machine.Transit(TCP_SEND)
	fmt.Printf("Current state is %s\n", machine.CurrentState().Name)
	fmt.Printf("\n\n")

	fmt.Printf("rcv SYN, expect next state to be SYN_RCVD\n")
	if !machine.HasTransition(TCP_RECV_SYN) {
		panic("state LISTEN cannot transit with TCP_RECV_SYN event")
	}
	r, _ = machine.GetResponse(TCP_RECV_SYN)
	fmt.Printf("Response to TCP_RECV_SYN event is %+v, ctrl flags should be %b\n", r, r.GetCtrlFlags())
	machine.Transit(TCP_RECV_SYN)
	fmt.Printf("Current state is %s\n", machine.CurrentState().Name)
	fmt.Printf("\n\n")

	fmt.Printf("rcv ACK, expect next state to be ESTAB\n")
	machine.Transit(TCP_RECV_ACK)
	fmt.Printf("Current state is %s\n", machine.CurrentState().Name)
	fmt.Printf("\n\n")

	fmt.Printf("CLOSE, expect next state to be FIN_WAIT_1\n")
	machine.Transit(TCP_CLOSE)
	fmt.Printf("Current state is %s\n", machine.CurrentState().Name)
	fmt.Printf("\n\n")

	fmt.Printf("rcv ACK, expect next state to be FIN_WAIT_2\n")
	machine.Transit(TCP_RECV_ACK)
	fmt.Printf("Current state is %s\n", machine.CurrentState().Name)
	fmt.Printf("\n\n")

	fmt.Printf("rcv FIN, expect next state to be TIME_WAIT\n")
	machine.Transit(TCP_RECV_FIN)
	fmt.Printf("Current state is %s\n", machine.CurrentState().Name)
	fmt.Printf("\n\n")
}
