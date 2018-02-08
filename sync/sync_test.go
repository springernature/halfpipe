package sync_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/springernature/halfpipe/sync"
	"github.com/blang/semver"
	"github.com/springernature/halfpipe"
)

type FakeGithubRelease struct {
	Version semver.Version
}

func (f FakeGithubRelease) GetLatestBinaryURL() (string, error) {
	panic("implement me")
}

func (f FakeGithubRelease) GetLatestVersion() (semver.Version, error) {
	getLatestVersionCalled = true
	return f.Version, nil
}

var getLatestVersionCalled bool

var _ = Describe("Sync", func() {

	BeforeEach(func() {
		getLatestVersionCalled = false
	})

	Context("Check", func() {
		It("returns empty error if using dev release and doesnt check github", func() {
			release := FakeGithubRelease{}
			sync := sync.Syncer{
				CurrentVersion: halfpipe.DevVersion,
				GithubRelease:  release,
			}
			Expect(sync.Check()).To(Not(HaveOccurred()))
			Expect(getLatestVersionCalled).To(BeFalse())
		})

		It("returns empty error if up to date", func() {
			release := FakeGithubRelease{Version: semver.Version{
				Major: 0,
				Minor: 0,
				Patch: 1,
			}}
			sync := sync.Syncer{
				CurrentVersion: semver.Version{
					Major: 0,
					Minor: 0,
					Patch: 1,
				},
				GithubRelease: release,
			}

			Expect(sync.Check()).To(Not(HaveOccurred()))
			Expect(getLatestVersionCalled).To(BeTrue())
		})

		It("returns error when the binary is out of date", func() {
			release := FakeGithubRelease{Version: semver.Version{
				Major: 0,
				Minor: 0,
				Patch: 2,
			}}
			sync := sync.Syncer{
				CurrentVersion: semver.Version{
					Major: 0,
					Minor: 0,
					Patch: 1,
				},
				GithubRelease: release,
			}
			Expect(sync.Check()).To(HaveOccurred())
			Expect(getLatestVersionCalled).To(BeTrue())
		})
	})

	Context("Update", func() {
		It("returns error when trying to upgrade dev release", func() {
			release := FakeGithubRelease{}
			sync := sync.Syncer{
				CurrentVersion: halfpipe.DevVersion,
				GithubRelease:  release,
			}
			Expect(sync.Update()).To(HaveOccurred())
		})
	})
})
