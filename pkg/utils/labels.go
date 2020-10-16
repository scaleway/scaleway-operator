package utils

// LabelsToTags transform labels into tags
func LabelsToTags(labels map[string]string) []string {
	tags := []string{}
	for key, value := range labels {
		tags = append(tags, key+"="+value)
	}
	return tags
}

// CompareTagsLabels returns true if the tags and labels are equal
func CompareTagsLabels(tags []string, labels map[string]string) bool {
	if len(tags) != len(labels) {
		return false
	}
	for key, value := range labels {
		found := false
		for _, tag := range tags {
			if tag == key+"="+value {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}
