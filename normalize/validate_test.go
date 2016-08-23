package normalize

import (
	"github.com/StackExchange/dnscontrol/models"
	"testing"
)

func Test_assert_no_enddot(t *testing.T) {
	var tests = []struct {
		experiment string
		isError    bool
	}{
		{"@", false},
		{"foo", false},
		{"foo.bar", false},
		{"foo.", true},
		{"foo.bar.", true},
	}

	for _, test := range tests {
		err := assert_no_enddot(test.experiment)
		checkError(t, err, test.isError, test.experiment)
	}
}

func checkError(t *testing.T, err error, shouldError bool, experiment string) {
	if err != nil && !shouldError {
		t.Errorf("%v: Error (%v)\n", experiment, err)
	}
	if err == nil && shouldError {
		t.Errorf("%v: Expected error but got none \n", experiment)
	}
}

func Test_assert_no_underscores(t *testing.T) {
	var tests = []struct {
		experiment string
		isError    bool
	}{
		{"@", false},
		{"foo", false},
		{"_foo", true},
		{"foo_", true},
		{"fo_o", true},
	}

	for _, test := range tests {
		err := assert_no_underscores(test.experiment)
		checkError(t, err, test.isError, test.experiment)
	}
}

func Test_assert_valid_ipv4(t *testing.T) {
	var tests = []struct {
		experiment string
		isError    bool
	}{
		{"1.2.3.4", false},
		{"1.2.3.4/10", true},
		{"1.2.3", true},
		{"foo", true},
	}

	for _, test := range tests {
		err := assert_valid_ipv4(test.experiment)
		checkError(t, err, test.isError, test.experiment)
	}
}

func Test_assert_valid_target(t *testing.T) {
	var tests = []struct {
		experiment string
		isError    bool
	}{
		{"@", false},
		{"foo", false},
		{"foo.bar.", false},
		{"foo.", false},
		{"foo.bar", true},
	}

	for _, test := range tests {
		err := assert_valid_target(test.experiment)
		checkError(t, err, test.isError, test.experiment)
	}
}

func Test_transform_cname(t *testing.T) {
	var tests = []struct {
		experiment string
		expected   string
	}{
		{"@", "old.com.new.com."},
		{"foo", "foo.old.com.new.com."},
		{"foo.bar", "foo.bar.old.com.new.com."},
		{"foo.bar.", "foo.bar.new.com."},
		{"chat.stackexchange.com.", "chat.stackexchange.com.new.com."},
	}

	for _, test := range tests {
		actual := transform_cname(test.experiment, "old.com", "new.com")
		if test.expected != actual {
			t.Errorf("%v: expected (%v) got (%v)\n", test.experiment, test.expected, actual)
		}
	}
}

func TestTransforms(t *testing.T) {
	var tests = []struct {
		givenIP         string
		expectedRecords []string
	}{
		{"0.0.5.5", []string{"2.0.5.5"}},
		{"3.0.5.5", []string{"5.5.5.5"}},
		{"7.0.5.5", []string{"9.9.9.9", "10.10.10.10"}},
	}
	const transform = "0.0.0.0~1.0.0.0~2.0.0.0~;   3.0.0.0~4.0.0.0~~5.5.5.5; 7.0.0.0~8.0.0.0~~9.9.9.9,10.10.10.10"
	for i, test := range tests {
		dc := &models.DomainConfig{
			Records: []*models.RecordConfig{
				{Type: "A", Target: test.givenIP, Metadata: map[string]string{"transform": transform}},
			},
		}
		err := applyRecordTransforms(dc)
		if err != nil {
			t.Errorf("error on test %d: %s", i, err)
			continue
		}
		if len(dc.Records) != len(test.expectedRecords) {
			t.Errorf("test %d: expect %d records but found %d", i, len(test.expectedRecords), len(dc.Records))
			continue
		}
		for r, rec := range dc.Records {
			if rec.Target != test.expectedRecords[r] {
				t.Errorf("test %d at index %d: records don't match. Expect %s but found %s.", i, r, test.expectedRecords[r], rec.Target)
				continue
			}
		}
	}
}