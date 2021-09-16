package node

import "time"

func (n *Node) pollingLoop() {
	for {
		select {
		case <-n.pollingExit:
			return
		default:
			if n.status != Connected {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			_, templateErr := n.GetBlockTemplate()
			if templateErr != nil {
				n.log.WithError(templateErr).Error("GetBlockTemplate")
				continue
			}
			n.GenerateWorkAsync(0)
		}
	}
}

func (n *Node) GenerateWorkAsync(removedTransactions int) {
	n.generateChan <- removedTransactions
}

func (n *Node) generateLoop() {
	var count int
	for {
		select {
		case <-n.generateExit:
			return
		case count = <-n.generateChan:
			var work = n.GenerateWork(count)
			if work != nil {
				n.workChan <- work
			}
		}
	}
}

func (n *Node) GenerateWork(removedTransactions int) *Work {
	block, blockErr := n.GetBlock(removedTransactions)
	if blockErr != nil {
		n.log.WithError(blockErr).Error("GetBlock")
	} else {
		work := NewWork(n, block)
		work.PlainHeader()
		return work
	}
	return nil
}
