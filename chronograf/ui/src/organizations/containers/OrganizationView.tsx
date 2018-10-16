// Libraries
import React, {PureComponent} from 'react'
import {WithRouterProps} from 'react-router'
import {connect} from 'react-redux'
import _ from 'lodash'

// Components
// import OrganizationViewContents from 'src/organizations/components/OrganizationViewContents'
import {Page} from 'src/pageLayout'
import ProfilePage from 'src/shared/components/profile_page/ProfilePage'
import Members from 'src/organizations/components/Members'

// Types
import {Organization, AppState} from 'src/types/v2'

// Decorators
import {ErrorHandling} from 'src/shared/decorators/errors'

interface StateProps {
  org: Organization
}

type Props = StateProps & WithRouterProps

@ErrorHandling
class OrganizationView extends PureComponent<Props> {
  public render() {
    const {org, params} = this.props

    return (
      <Page>
        <Page.Header fullWidth={false}>
          <Page.Header.Left>
            <Page.Title title="Organization" />
          </Page.Header.Left>
          <Page.Header.Right />
        </Page.Header>
        <Page.Contents fullWidth={false} scrollable={true}>
          <div className="col-xs-12">
            <ProfilePage
              name={org.name}
              parentUrl="/organizations"
              activeTabUrl={params.tab}
            >
              <ProfilePage.Section
                id="org-view-tab--members"
                url="members"
                title="Members"
              >
                <Members />
              </ProfilePage.Section>
              <ProfilePage.Section
                id="org-view-tab--buckets"
                url="buckets"
                title="Buckets"
              >
                <div>Render bucket memes here</div>
              </ProfilePage.Section>
              <ProfilePage.Section
                id="org-view-tab--dashboards"
                url="dashboards"
                title="Dashboards"
              >
                <div>Render dashboard memes here</div>
              </ProfilePage.Section>
              <ProfilePage.Section
                id="org-view-tab--tasks"
                url="tasks"
                title="Tasks"
              >
                <div>Render dashboard memes here</div>
              </ProfilePage.Section>
            </ProfilePage>
          </div>
        </Page.Contents>
      </Page>
    )
  }
}

const mstp = (state: AppState, props: WithRouterProps) => {
  const {orgs} = state

  const org = orgs.find(o => o.id === props.params.orgID)

  return {
    org,
  }
}

export default connect<StateProps, {}, {}>(mstp, null)(OrganizationView)
