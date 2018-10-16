// Libraries
import React, {PureComponent} from 'react'

// Types
import {Member} from 'src/types/v2'

interface Props {
  members: Member[]
}

export default class Members extends PureComponent<Props> {
  public render() {
    return <div>{JSON.stringify(this.props.members)}</div>
  }
}
