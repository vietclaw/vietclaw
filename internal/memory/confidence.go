package memory

func confidenceValue(conf Confidence) float64 {
	switch conf {
	case ConfidenceTemporary:
		return 0.35
	case ConfidenceInferred:
		return 0.7
	default:
		return 1.0
	}
}

func confidenceLabel(value float64) Confidence {
	if value < 0.5 {
		return ConfidenceTemporary
	}
	if value < 0.9 {
		return ConfidenceInferred
	}
	return ConfidenceConfirmed
}
