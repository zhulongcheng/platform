// Libraries
import React, {PureComponent} from 'react'
import {WithRouterProps} from 'react-router'
import {connect} from 'react-redux'
import _ from 'lodash'

// Components
import {Page} from 'src/pageLayout'
import CreateOrgOverlay from 'src/organizations/components/CreateOrgOverlay'
import OrganizationsIndexContents from 'src/organizations/components/OrganizationsIndexContents'
import {
  Button,
  IconFont,
  ComponentColor,
  OverlayTechnology,
} from 'src/clockface'

// Actions
import {createOrg} from 'src/organizations/actions'

// Types
import {Organization, Links} from 'src/types/v2'

// Decorators
import {ErrorHandling} from 'src/shared/decorators/errors'

interface StateProps {
  links: Links
  orgs: Organization[]
}

interface DispatchProps {
  onCreateOrg: typeof createOrg
}

interface State {
  modalState: ModalState
}

enum ModalState {
  Open = 'open',
  Closed = 'closed',
}

type Props = StateProps & DispatchProps & WithRouterProps

@ErrorHandling
class OrganizationsIndex extends PureComponent<Props, State> {
  constructor(props) {
    super(props)
    this.state = {
      modalState: ModalState.Closed,
    }
  }
  public render() {
    const {orgs, links, onCreateOrg} = this.props
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
            onCreateOrg={onCreateOrg}
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

const mstp = (state): StateProps => {
  const {orgs, links} = state

  return {
    orgs,
    links,
  }
}

const mdtp: DispatchProps = {
  onCreateOrg: createOrg,
}

export default connect<StateProps, DispatchProps>(
  mstp,
  mdtp
)(OrganizationsIndex)
