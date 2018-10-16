// Libraries
import React, {PureComponent} from 'react'
import {InjectedRouter} from 'react-router'
import {connect} from 'react-redux'
import _ from 'lodash'

// Components
import OrganizationsIndexContents from 'src/organizations/components/OrganizationsIndexContents'
import {Page} from 'src/pageLayout'
import {Button, ComponentColor, IconFont} from 'src/clockface'

// Types
import {Organization, Links} from 'src/types/v2'

// Decorators
import {ErrorHandling} from 'src/shared/decorators/errors'

interface Props {
  router: InjectedRouter
  links: Links
  orgs: Organization[]
}

@ErrorHandling
class OrganizationsIndex extends PureComponent<Props> {
  public render() {
    const {orgs} = this.props

    return (
      <Page>
        <Page.Header fullWidth={false}>
          <Page.Header.Left>
            <Page.Title title="Organizations" />
          </Page.Header.Left>
          <Page.Header.Right>
            <Button
              color={ComponentColor.Primary}
              onClick={this.handleCreateOrg}
              icon={IconFont.Plus}
              text="Create Organization"
              titleText="Create a new Organization"
            />
          </Page.Header.Right>
        </Page.Header>
        <Page.Contents fullWidth={false} scrollable={true}>
          <OrganizationsIndexContents
            orgs={orgs}
            onDeleteOrg={this.handleDeleteOrg}
          />
        </Page.Contents>
      </Page>
    )
  }

  private handleCreateOrg = (): void => {
    console.log('make a new org')
  }

  private handleDeleteOrg = (org: Organization): void => {
    console.log('delete organization with id ', org.id)
  }
}

const mstp = state => {
  const {orgs, links} = state

  return {
    orgs,
    links,
  }
}

export default connect(mstp)(OrganizationsIndex)
