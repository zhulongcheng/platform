// Libraries
import React, {PureComponent} from 'react'

// Components
import IndexList from 'src/shared/components/index_views/IndexList'
import {Alignment, ComponentSize, EmptyState} from 'src/clockface'

// Types
import {Bucket} from 'src/types/v2'
import {
  IndexListColumn,
  IndexListRow,
} from 'src/shared/components/index_views/IndexListTypes'

interface Props {
  buckets: Bucket[]
}

// Decorators
import {ErrorHandling} from 'src/shared/decorators/errors'

@ErrorHandling
export default class Buckets extends PureComponent<Props> {
  public render() {
    return (
      <IndexList
        columns={this.columns}
        rows={this.rows}
        emptyState={this.emptyState}
      />
    )
  }

  private get columns(): IndexListColumn[] {
    return [
      {
        key: 'bucket--name',
        title: 'Name',
        size: 500,
        showOnHover: false,
        align: Alignment.Left,
      },
      {
        key: 'bucket--retention',
        title: 'Retention Period',
        size: 100,
        showOnHover: false,
        align: Alignment.Left,
      },
    ]
  }

  private get rows(): IndexListRow[] {
    const {buckets} = this.props

    return buckets.map(bucket => ({
      disabled: false,
      columns: [
        {
          key: 'bucket--name',
          contents: <a href="#">{bucket.name}</a>,
        },
        {
          key: 'bucket--retention',
          contents: bucket.retentionPeriod,
        },
      ],
    }))
  }

  private get emptyState(): JSX.Element {
    return (
      <EmptyState size={ComponentSize.Large}>
        <EmptyState.Text text="Oh noes I dun see na buckets" />
      </EmptyState>
    )
  }
}
