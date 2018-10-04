// Libraries
import React, {PureComponent, ReactChildren} from 'react'
import {Provider, Subscribe} from 'unstated'

import Nav from 'src/page_layout'
import {LinksContainer} from 'src/LinksContainer'
import Notifications from 'src/shared/components/notifications/Notifications'

import {RemoteDataState} from 'src/types'

const links = new LinksContainer()

interface Props {
  linksContainer: LinksContainer
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
    await this.props.linksContainer.getLinks()
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
    <Provider inject={[links]}>
      <Subscribe to={[LinksContainer]}>
        {(linksContainer: LinksContainer) => (
          <App {...props} linksContainer={linksContainer} />
        )}
      </Subscribe>
    </Provider>
  )
}

export default ConnectedApp
