package cmdr

import (
	"fmt"
	"testing"
	"time"
)

func statEq(l, r *State) (res bool, msg string) {
	if l.Name != r.Name {
		return false, fmt.Sprintf("name mismatch: '%s' / '%s'", l.Name, r.Name)
	}
	if l.Jumps.Len() != r.Jumps.Len() {
		return false, fmt.Sprintf("# jumps in hist differs: %d / %d", l.Jumps.Len(), r.Jumps.Len())
	}
	if l.Jumps.WrPtr != r.Jumps.WrPtr {
		return false, fmt.Sprintf("# wrptr in hist differs: %d / %d", l.Jumps.WrPtr, r.Jumps.WrPtr)
	}
	for i := range l.Jumps.Hist {
		lj, rj := &l.Jumps.Hist[i], &r.Jumps.Hist[i]
		if lj.SysId != rj.SysId {
			return false, fmt.Sprintf("different sys-ids in jump @%d: %d / %d",
				i, lj.SysId, rj.SysId)
		}
		if lj.When != rj.When {
			return false, fmt.Sprintf("different t in jump @%d: %s / %s",
				i, lj.When, rj.When)
		}
	}
	return true, ""
}

func TestJumpHist(t *testing.T) {
	var jh JumpHist
	for i := 0; i < 2*JumpHistLen+3; i++ {
		jh.Add(1, time.Now())
		if i < JumpHistLen {
			if jh.Len() != i+1 {
				t.Fatalf("Step %d: unexpected len=%d, want %d", i, jh.Len(), i+1)
			}
		} else if jh.Len() != JumpHistLen {
			t.Fatalf("Step %d: unexpected len=%d, want %d", i, jh.Len(), JumpHistLen)
		}
	}
}

func TestState_safeload(t *testing.T) {
	s1 := NewState(nil)
	s1.Name = "Jameson"
	if s1.Jumps.Len() != 0 {
		t.Fatal("jh must be empty")
	}
	s1.Jumps.Add(1, time.Now().Add(0*time.Second))
	s1.Jumps.Add(2, time.Now().Add(1*time.Second))
	s1.Jumps.Add(3, time.Now().Add(2*time.Second))
	if s1.Jumps.Len() != 3 {
		t.Fatalf("jh must hav elength 3, got %d", s1.Jumps.Len())
	}
	fname := t.Name() + ".state"
	err := s1.Save(fname)
	if err != nil {
		t.Fatal(err)
	}
	s2 := NewState(nil)
	err = s2.Load(fname)
	if err != nil {
		t.Fatal(err)
	}
	if eq, msg := statEq(s1, s2); !eq {
		t.Error(msg)
	}
}
