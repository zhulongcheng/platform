// Libraries
import React, {PureComponent} from 'react'
import {Link} from 'react-router'

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

export default class OrganizationList extends PureComponent<Props, State> {
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

    if (this.isEmpty) {
      return <div>Looks like you dont have any organizations</div>
    }

    return (
      <>
        <h4>Organizations</h4>
        <ul>
          {orgs.map(o => (
            <li>
              <Link to={`organization/${o.id}`}>{o.name}</Link>
            </li>
          ))}
        </ul>
      </>
    )
  }

  get isEmpty() {
    return !this.state.orgs.length
  }
}
