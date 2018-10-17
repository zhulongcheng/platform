// Libraries
import React, {PureComponent} from 'react'

// Types
import {Organization} from 'src/types/v2'

interface Props {
  org: Organization
}

export default class OrgOptions extends PureComponent<Props> {
  public render() {
    return <div>{JSON.stringify(this.props.org, null, 2)}</div>
  }
}
