// Libraries
import React, {PureComponent, ReactChildren} from 'react'
import {Subscribe} from 'unstated'

import Nav from 'src/page_layout'
import {LinksContainer} from 'src/LinksContainer'
import Notifications from 'src/shared/components/notifications/Notifications'

import {RemoteDataState} from 'src/types'

interface Props {
  getLinks: LinksContainer['getLinks']
  children: ReactChildren
}

interface State {
  loading: RemoteDataState
}

class App extends PureComponent<Props, State> {
  public state = {
    loading: RemoteDataState.NotStarted,
  }

  public async componentDidMount() {
    await this.props.getLinks()
    this.setState({loading: RemoteDataState.Done})
  }

  public render() {
    const {children} = this.props

    if (this.isLoading) {
      return (
        <div className="chronograf-root">
          <div className="page-spinner" />
        </div>
      )
    }

    return (
      <div className="chronograf-root">
        <Notifications />
        <Nav />
        {children}
      </div>
    )
  }

  get isLoading(): boolean {
    const {loading} = this.state
    return (
      loading === RemoteDataState.Loading ||
      loading === RemoteDataState.NotStarted
    )
  }
}

const ConnectedApp = (props: Props) => {
  return (
    <Subscribe to={[LinksContainer]}>
      {(linksContainer: LinksContainer) => (
        <App {...props} getLinks={linksContainer.getLinks} />
      )}
    </Subscribe>
  )
}

export default ConnectedApp
