package app

import (
	"container/heap"
	"os/exec"
	"regexp"
	"time"
)

const ChanJEvt = "jevt"

type SpeakCfg struct {
	Cmd     string
	Flags   []string
	Old     time.Duration
	ChanCfg map[string]*ChanConfig
}

func (sc *SpeakCfg) init() error {
	for _, chc := range sc.ChanCfg {
		if err := chc.parseRegexs(); err != nil {
			return err
		}
	}
	return nil
}

type ChanConfig struct {
	Flags []string
	Black []string
	White []string
	bls   []*regexp.Regexp
	wls   []*regexp.Regexp
}

func (chc *ChanConfig) parseRegexs() error {
	chc.bls = nil
	chc.wls = nil
	for _, bstr := range chc.Black {
		rx, err := regexp.Compile(bstr)
		if err != nil {
			return err
		}
		chc.bls = append(chc.bls, rx)
	}
	for _, wstr := range chc.White {
		rx, err := regexp.Compile(wstr)
		if err != nil {
			return err
		}
		chc.wls = append(chc.wls, rx)
	}
	return nil
}

func (cc *ChanConfig) Filter(msg string) bool {
	res := true
	for i, f := range cc.bls {
		if f.MatchString(msg) {
			log.Tracea("speak `filter` block `msg`", cc.Black[i], msg)
			res = false
			break
		}
	}
	if !res {
		for i, f := range cc.wls {
			if f.MatchString(msg) {
				log.Tracea("speak `filter` selects `msg`", cc.White[i], msg)
				res = true
				break
			}
		}
	}
	return res
}

type VoiceMsg struct {
	Chan string
	Prio int
	Txt  string
}

type vMsgQ []VoiceMsg

func (q vMsgQ) Len() int { return len(q) }

func (q vMsgQ) Less(i, j int) bool { return q[i].Prio > q[j].Prio }

func (q vMsgQ) Swap(i, j int) { q[i], q[j] = q[j], q[i] }

func (q *vMsgQ) Push(m interface{}) { *q = append(*q, m.(VoiceMsg)) }

func (q *vMsgQ) Pop() (m interface{}) {
	l := len(*q) - 1
	m = (*q)[l]
	*q = (*q)[:l]
	return m
}

func (spk *SpeakCfg) run() chan<- VoiceMsg {
	if spk.Cmd == "" {
		log.Warns("speak not configured")
		return nil
	}
	ch := make(chan VoiceMsg, 16)
	go spk.speaker(ch)
	return ch
}

func (spk *SpeakCfg) speaker(msgs <-chan VoiceMsg) {
	log.Infoa("running `speaker` with `flags`", spk.Cmd, spk.Flags)
	var mq vMsgQ
	for msg := range msgs {
		heap.Push(&mq, msg)
		for len(mq) > 0 {
			for len(msgs) > 0 {
				heap.Push(&mq, <-msgs)
			}
			msg = heap.Pop(&mq).(VoiceMsg)
			argv := spk.Flags
			ccfg := spk.ChanCfg[msg.Chan]
			if ccfg == nil {
				ccfg = spk.ChanCfg[""]
			}
			if ccfg == nil {
				ccfg = &ChanConfig{}
			} else if !ccfg.Filter(msg.Txt) {
				log.Debuga("speak filter dropped `channel` `message`", msg.Chan, msg.Txt)
				continue
			}
			log.Debuga("speak on `channel` with `prio` `text`", msg.Chan, msg.Prio, msg.Txt)
			if len(ccfg.Flags) > 0 {
				argv = append(argv, ccfg.Flags...)
			}
			argv = append(argv, msg.Txt)
			cmd := exec.Command(spk.Cmd, argv...)
			if err := cmd.Run(); err != nil {
				log.Errora("speak `err`", err)
			}
		}
		log.Debugs("speak backlog empty, wait for new voice messages")
	}
	log.Infos("stop speaker")
}
