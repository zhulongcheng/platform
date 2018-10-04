// Libraries
import React, {PureComponent} from 'react'
import DashboardList from 'src/user/components/DashboardsList'

// APIs
import {getDashboards} from 'src/dashboards/apis/v2'

// Types
import {Dashboard, RemoteDataState} from 'src/types'

interface State {
  loading: RemoteDataState
  dashboards: Dashboard[]
}

interface Props {
  dashboardsLink: string
}

export default class UserDashboards extends PureComponent<Props, State> {
  constructor(props) {
    super(props)
    this.state = {
      dashboards: [],
      loading: RemoteDataState.NotStarted,
    }
  }

  public async componentDidMount() {
    const {dashboardsLink} = this.props
    const dashboards = await getDashboards(dashboardsLink)
    this.setState({dashboards, loading: RemoteDataState.Done})
  }

  public render() {
    const {dashboards, loading} = this.state
    if (loading === RemoteDataState.Loading) {
      return <div> Loading...</div>
    }

    return <DashboardList dashboards={dashboards} />
  }
}
