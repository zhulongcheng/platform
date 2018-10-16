// Libraries
import React, {PureComponent} from 'react'

// Types
import {Dashboard} from 'src/types/v2'

// Decorators
import {ErrorHandling} from 'src/shared/decorators/errors'

interface Props {
  dashboards: Dashboard[]
}

@ErrorHandling
export default class Dashboards extends PureComponent<Props> {
  public render() {
    return <div>{JSON.stringify(this.props.dashboards)}</div>
  }
}
