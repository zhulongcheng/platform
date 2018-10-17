// Libraries
import React, {PureComponent} from 'react'
import {InjectedRouter} from 'react-router'
import {connect} from 'react-redux'
import _ from 'lodash'

// Components
import OrganizationsIndexContents from 'src/organizations/components/OrganizationsIndexContents'
import {Page} from 'src/pageLayout'
import {OverlayTechnology} from 'src/clockface'
import CreateOrgOverlay from 'src/organizations/components/CreateOrgOverlay'
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

interface State {
  modalState: ModalState
}

enum ModalState {
  Open = 'open',
  Closed = 'closed',
}

@ErrorHandling
class OrganizationsIndex extends PureComponent<Props, State> {
  constructor(props) {
    super(props)
    this.state = {
      modalState: ModalState.Closed,
    }
  }
  public render() {
    const {orgs, links} = this.props
    const {modalState} = this.state

    return (
      <>
        <Page>
          <Page.Header fullWidth={false}>
            <Page.Header.Left>
              <Page.Title title="Organizations" />
            </Page.Header.Left>
            <Page.Header.Right>
              <Button
                color={ComponentColor.Primary}
                onClick={this.handleOpenModal}
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
        <OverlayTechnology visible={modalState === ModalState.Open}>
          <CreateOrgOverlay
            link={links.orgs}
            onCloseModal={this.handleCloseModal}
          />
        </OverlayTechnology>
      </>
    )
  }

  private handleOpenModal = (): void => {
    this.setState({modalState: ModalState.Open})
  }

  private handleCloseModal = (): void => {
    this.setState({modalState: ModalState.Closed})
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
