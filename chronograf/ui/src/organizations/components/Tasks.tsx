// Libraries
import React, {PureComponent} from 'react'

// Types
import {Task} from 'src/types/v2'

interface Props {
  tasks: Task[]
}

export default class Members extends PureComponent<Props> {
  public render() {
    return <div>{JSON.stringify(this.props.tasks)}</div>
  }
}
