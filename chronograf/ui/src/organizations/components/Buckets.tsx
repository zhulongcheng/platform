// Libraries
import React, {PureComponent} from 'react'

// Types
import {Bucket} from 'src/types/v2'

interface Props {
  buckets: Bucket[]
}

export default class Buckets extends PureComponent<Props> {
  public render() {
    return <div>{JSON.stringify(this.props.buckets)}</div>
  }
}
