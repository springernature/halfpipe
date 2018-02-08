package sync_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/springernature/halfpipe/sync"
	"github.com/blang/semver"
)

type FakeGithubRelease struct {
	Version semver.Version
}

func (f FakeGithubRelease) GetLatestBinaryURL() (string, error) {
	panic("implement me")
}

func (f FakeGithubRelease) GetLatestVersion() (semver.Version, error) {
	return f.Version, nil
}

var _ = Describe("Sync", func() {

	Context("Check", func() {
		It("returns empty error if up to date", func() {
			sync := sync.Syncer{
				CurrentVersion: semver.Version{
					Major: 0,
					Minor: 0,
					Patch: 1,
				},
				GithubRelease: FakeGithubRelease{Version: semver.Version{
					Major: 0,
					Minor: 0,
					Patch: 1,
				}},
			}

			Expect(sync.Check()).To(Not(HaveOccurred()))
		})

		It("returns error when the binary is out of date", func() {
			sync := sync.Syncer{
				CurrentVersion: semver.Version{
					Major: 0,
					Minor: 0,
					Patch: 1,
				},
				GithubRelease: FakeGithubRelease{Version: semver.Version{
					Major: 0,
					Minor: 0,
					Patch: 2,
				}},
			}
			Expect(sync.Check()).To(HaveOccurred())
		})
	})
})
