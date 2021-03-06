package v1alpha1

func (g *Genesis) Default() {
	if g.NetworkID == 0 {
		g.NetworkID = DefaultNetworkID
	}

	if g.ChainID == 0 {
		g.ChainID = DefaultChainID
	}

	if g.QBFT.EpochLength == 0 {
		g.QBFT.EpochLength = DefaultQBFTEpochLength
	}

	if g.QBFT.BlockPeriodSeconds == 0 {
		g.QBFT.BlockPeriodSeconds = DefaultQBFTBlockPeriodSeconds
	}

	if g.QBFT.RequestTimeoutSeconds == 0 {
		g.QBFT.RequestTimeoutSeconds = DefaultQBFTRequestTimeoutSeconds
	}
}
