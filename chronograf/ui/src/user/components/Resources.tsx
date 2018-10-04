// Libraries
import React, {PureComponent} from 'react'
import {Subscribe} from 'unstated'

// Containers
import {LinksContainer} from 'src/LinksContainer'

// Components
import Orgs from 'src/user/components/Orgs'
import Dashboards from 'src/user/components/Dashboards'
import Support from 'src/user/components/Support'
import Settings from 'src/user/components/Settings'
import Avatar from 'src/shared/components/avatar/Avatar'
import {Panel} from 'src/clockface'

// Types
import {Links} from 'src/types/v2'
import {User} from 'src/types/v2/user'

interface Props {
  user: User
}

interface ConnectedProps {
  links: Links
}

class ResourceLists extends PureComponent<ConnectedProps & Props> {
  public render() {
    const {links, user} = this.props

    return (
      <>
        <Panel>
          <Panel.Header title="My Account">
            <Avatar imageURI={user.avatar} diameterPixels={45} />
          </Panel.Header>
          <Panel.Body>
            <Settings />
          </Panel.Body>
        </Panel>
        <Panel>
          <Panel.Header title="Organizations">
            <button>Create</button>
          </Panel.Header>
          <Panel.Body>
            <Orgs orgsLink={links.orgs} />
          </Panel.Body>
        </Panel>
        <Panel>
          <Panel.Header title="Dashboards">
            <button>Create</button>
          </Panel.Header>
          <Panel.Body>
            <Dashboards dashboardsLink={links.dashboards} />
          </Panel.Body>
        </Panel>
        <Panel>
          <Panel.Header title="Useful Links" />
          <Panel.Body>
            <Support />
          </Panel.Body>
        </Panel>
      </>
    )
  }
}

const ConnectedResourceLists = props => {
  return (
    <Subscribe to={[LinksContainer]}>
      {(linksContainer: LinksContainer) => {
        return <ResourceLists {...props} links={linksContainer.state.links} />
      }}
    </Subscribe>
  )
}

export default ConnectedResourceLists
