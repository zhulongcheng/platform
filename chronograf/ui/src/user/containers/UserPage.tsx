// Libraries
import React, {PureComponent} from 'react'

// Components
import {Page} from 'src/page_layout'
import ProfilePage from 'src/shared/components/profile_page/ProfilePage'
import UserSettings from 'src/user/components/UserSettings'
import TokenManager from 'src/user/components/TokenManager'
import Resources from 'src/user/components/Resources'
import Header from 'src/user/components/UserPageHeader'

// Types
import {Organization, Dashboard} from 'src/types'

// Decorators
import {ErrorHandling} from 'src/shared/decorators/errors'

// MOCK DATA
import {LeroyJenkins, Orgs} from 'src/user/mockUserData'

interface UserToken {
  id: string
  name: string
  secretKey: string
}

interface User {
  id: string
  name: string
  email: string
  avatar: string
  tokens: UserToken[]
}

interface Props {
  user: User
  organizations: Organization[]
  dashboards: Array<Partial<Dashboard>>
  params: {
    tab: string
  }
}

@ErrorHandling
export class UserPage extends PureComponent<Props> {
  public static defaultProps: Partial<Props> = {
    user: LeroyJenkins,
    organizations: Orgs,
  }

  public render() {
    const {user, params} = this.props

    return (
      <Page>
        <Header title={`Howdy, ${user.name}!`} />
        <Page.Contents fullWidth={false} scrollable={true}>
          <div className="col-xs-7">
            <ProfilePage
              name={user.name}
              avatar={user.avatar}
              parentUrl="/user"
              activeTabUrl={params.tab}
            >
              <ProfilePage.Section
                id="user-profile-tab--settings"
                url="settings"
                title="Settings"
              >
                <UserSettings blargh="User Settings" />
              </ProfilePage.Section>
              <ProfilePage.Section
                id="user-profile-tab--tokens"
                url="tokens"
                title="Tokens"
              >
                <TokenManager token="Token Manager" />
              </ProfilePage.Section>
            </ProfilePage>
          </div>
          <div className="col-xs-5">
            <Resources />
          </div>
        </Page.Contents>
      </Page>
    )
  }
}

export default UserPage
