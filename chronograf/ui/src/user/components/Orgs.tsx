// Libraries
import React, {PureComponent} from 'react'

// Components
import OrgsList from 'src/user/components/OrgsList'

// APIs
import {getOrgs} from 'src/organizations/apis'

// Types
import {Organization, RemoteDataState} from 'src/types'

interface Props {
  orgsLink: string
}

interface State {
  orgs: Organization[]
  loading: RemoteDataState
}

export default class Orgs extends PureComponent<Props, State> {
  constructor(props) {
    super(props)
    this.state = {
      orgs: [],
      loading: RemoteDataState.NotStarted,
    }
  }

  public async componentDidMount() {
    const {orgsLink} = this.props
    const {orgs} = await getOrgs(orgsLink)
    this.setState({orgs, loading: RemoteDataState.Done})
  }

  public render() {
    const {loading, orgs} = this.state
    if (loading === RemoteDataState.Loading) {
      return <div> Loading...</div>
    }

    return <OrgsList orgs={orgs} />
  }

  get isEmpty() {
    return !this.state.orgs.length
  }
}
