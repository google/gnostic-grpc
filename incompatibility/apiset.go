package incompatibility

func NewAnalysis() *ApiSetIncompatibility {
	incompatibilityMap := make(map[string]*IncompatibilityAnalysis)
	for _, incompClassString := range IncompatibiltiyClassification_name {
		incompatibilityMap[incompClassString] =
			&IncompatibilityAnalysis{FilesWithIncompatibility: 0}
	}
	return &ApiSetIncompatibility{
		OpenApiFiles:               0,
		IncompatibleFiles:          0,
		AnalysisPerIncompatibility: incompatibilityMap,
	}
}

func FileReport2Analysis(reports ...*IncompatibilityReport) *ApiSetIncompatibility {
	analysis := NewAnalysis()
	for _, report := range reports {
		AggregateIncompatibilities(analysis, report.GetIncompatibilities()...)
	}
	return analysis
}

// TODO
func AggregateAnalysis(analysis ...*ApiSetIncompatibility) *ApiSetIncompatibility {
	return nil
}

// TODO
func AggregateIncompatibilities(analysis *ApiSetIncompatibility, incompatibilities ...*Incompatibility) {
	for _, incompatibility := range incompatibilities {
		addIncompatibility2Analysis(analysis, incompatibility)
	}
}

func addIncompatibility2Analysis(analysis *ApiSetIncompatibility, incompatibility *Incompatibility) {
	analysis.OpenApiFiles++
	// if analysis.AnalysisPerIncompatibility[incompatibility.Classification.String()]++
}
