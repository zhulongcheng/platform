// Libraries
import React, {PureComponent} from 'react'

// Components
import {Page} from 'src/page_layout'
import Resources from 'src/user/components/Resources'
import Header from 'src/user/components/UserPageHeader'

// Types
import {Organization, Dashboard} from 'src/types'
import {User} from 'src/types/v2/user'

// Decorators
import {ErrorHandling} from 'src/shared/decorators/errors'

// MOCK DATA
import {LeroyJenkins, Orgs} from 'src/user/mockUserData'

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
    const {user} = this.props

    return (
      <Page>
        <Header title={`Howdy, ${user.name}!`} />
        <Page.Contents fullWidth={false} scrollable={true}>
          <div className="col-xs-7" />
          <div className="col-xs-5">
            <Resources user={user} />
          </div>
        </Page.Contents>
      </Page>
    )
  }
}

export default UserPage
