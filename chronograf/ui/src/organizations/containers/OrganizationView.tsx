// Libraries
import React, {PureComponent} from 'react'
import {WithRouterProps} from 'react-router'
import {connect} from 'react-redux'
import _ from 'lodash'

// Components
// import OrganizationViewContents from 'src/organizations/components/OrganizationViewContents'
import {Page} from 'src/pageLayout'

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
    const {org} = this.props

    return (
      <Page>
        <Page.Header fullWidth={false}>
          <Page.Header.Left>
            <Page.Title title={org.name} />
          </Page.Header.Left>
          <Page.Header.Right />
        </Page.Header>
        <Page.Contents fullWidth={false} scrollable={true}>
          {/* <OrganizationViewContents
            org={org}
            onDeleteOrg={this.handleDeleteOrg}
          /> */}
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
