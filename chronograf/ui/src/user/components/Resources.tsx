// Libraries
import React, {PureComponent} from 'react'
import {Subscribe} from 'unstated'

// Containers
import {LinksContainer} from 'src/LinksContainer'

// Components
import Orgs from 'src/user/components/Orgs'
import Dashboards from 'src/user/components/Dashboards'
import Support from 'src/user/components/Support'
import {Panel} from 'src/clockface'

// Types
import {Links} from 'src/types/v2'

interface ConnectedProps {
  links: Links
}

class ResourceLists extends PureComponent<ConnectedProps> {
  public render() {
    const {links} = this.props

    return (
      <>
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
