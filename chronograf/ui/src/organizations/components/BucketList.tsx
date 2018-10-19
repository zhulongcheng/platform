// Libraries
import React, {PureComponent} from 'react'

// Components
import {
  IndexListColumn,
  IndexListRow,
} from 'src/shared/components/index_views/IndexListTypes'
import IndexList from 'src/shared/components/index_views/IndexList'
import {Alignment} from 'src/clockface'

// Types
import {Bucket} from 'src/types/v2'

interface PrettyBucket extends Bucket {
  retentionPeriod: string
}

interface Props {
  buckets: PrettyBucket[]
  emptyState: any
}

export default class BucketList extends PureComponent<Props> {
  public render() {
    return (
      <IndexList
        rows={this.rows}
        columns={this.columns}
        emptyState={this.props.emptyState}
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
        size: 200,
        showOnHover: false,
        align: Alignment.Right,
      },
    ]
  }

  private get rows(): IndexListRow[] {
    return this.props.buckets.map(bucket => ({
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
}
