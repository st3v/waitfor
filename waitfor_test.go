package waitfor_test

import (
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/st3v/waitfor"
	"golang.org/x/net/context"
)

type condition struct {
	sync.RWMutex
	counter int
	result  bool
}

func (c *condition) Check() bool {
	c.Lock()
	defer c.Unlock()
	c.counter++
	return c.result
}

func (c *condition) CheckCount() int {
	c.RLock()
	defer c.RUnlock()
	return c.counter
}

func (c *condition) SetResult(result bool) {
	c.Lock()
	defer c.Unlock()
	c.result = result
}

var _ = Describe("waitfor", func() {
	var (
		cond     condition
		interval = 10 * time.Millisecond
	)

	BeforeEach(func() {
		cond.Lock()
		cond = condition{}
	})

	Describe(".ConditionWithTimeout", func() {
		var (
			err     error
			timeout = 100 * time.Millisecond
		)

		JustBeforeEach(func() {
			err = waitfor.ConditionWithTimeout(cond.Check, interval, timeout)
		})

		Context("when the check does not succeed", func() {
			It("returns a corresponding error", func() {
				Expect(err).To(MatchError(waitfor.ErrTimeoutExceeded))
			})

			It("returns after the specified timeout", func() {
				Expect(cond.CheckCount()).To(BeNumerically("<", 11))
			})

			It("repeatedly checked the condition", func() {
				Expect(cond.CheckCount()).To(BeNumerically(">", 3))
			})
		})

		Context("when the check succeeds initially", func() {
			BeforeEach(func() {
				cond.SetResult(true)
			})

			It("checked the condition only once", func() {
				Expect(cond.CheckCount()).To(Equal(1))
			})

			It("does not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("when the check succeeds eventually", func() {
			BeforeEach(func() {
				go func() {
					<-time.After(timeout / 2)
					cond.SetResult(true)
				}()
			})

			It("checked the condition multiple times", func() {
				Expect(cond.CheckCount()).To(BeNumerically(">", 1))
			})

			It("does not return an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe(".Condition", func() {
		var (
			ctx     context.Context
			cancel  context.CancelFunc
			errChan chan error
		)

		JustBeforeEach(func() {
			ctx, cancel = context.WithCancel(context.Background())
			errChan = make(chan error)
			go waitfor.Condition(cond.Check, interval, errChan, ctx)
		})

		AfterEach(func() {
			cancel()
		})

		Context("when the check does not succeed", func() {
			It("keeps checking the condition repeatedly", func() {
				Eventually(cond.CheckCount).Should(BeNumerically(">=", 20))
			})

			Context("when the context is canceled", func() {
				It("stops checking the condition", func() {
					cancel()
					Eventually(errChan).Should(Receive())

					count := cond.CheckCount()
					Consistently(cond.CheckCount).Should(Equal(count))
				})

				It("returns a corresponding error on its error channel", func() {
					cancel()
					Eventually(errChan).Should(Receive(MatchError(context.Canceled)))
				})
			})
		})

		Context("when the check succeeds", func() {
			BeforeEach(func() {
				cond.SetResult(true)
			})

			It("checks the condition only once", func() {
				Eventually(cond.CheckCount).Should(Equal(1))
				Consistently(cond.CheckCount).Should(Equal(1))
			})
		})
	})
})
