// Libraries
import React, {Component} from 'react'
import {Link} from 'react-router'
import _ from 'lodash'

// Components
import IndexList from 'src/shared/components/index_views/IndexList'
import DeleteOrgButton from 'src/organizations/components/DeleteOrgButton'
import {
  ComponentSpacer,
  Alignment,
  ComponentSize,
  EmptyState,
} from 'src/clockface'

// Decorators
import {ErrorHandling} from 'src/shared/decorators/errors'

// Types
import {Organization} from 'src/types/v2'
import {
  IndexListColumn,
  IndexListRow,
} from 'src/shared/components/index_views/IndexListTypes'

interface Props {
  orgs: Organization[]
  onDeleteOrg: (org: Organization) => void
}

@ErrorHandling
class OrganizationsPageContents extends Component<Props> {
  public render() {
    return (
      <div className="col-md-12">
        <IndexList
          columns={this.columns}
          rows={this.rows}
          emptyState={this.emptyState}
        />
      </div>
    )
  }

  private get columns(): IndexListColumn[] {
    return [
      {
        key: 'organization--name',
        title: 'Name',
        size: 500,
        showOnHover: false,
        align: Alignment.Left,
      },
      {
        key: 'organization--membership',
        title: '',
        size: 100,
        showOnHover: false,
        align: Alignment.Center,
      },
      {
        key: 'organization--actions',
        title: '',
        size: 200,
        showOnHover: true,
        align: Alignment.Right,
      },
    ]
  }

  private get rows(): IndexListRow[] {
    const {orgs, onDeleteOrg} = this.props

    return orgs.map(o => ({
      disabled: false,
      columns: [
        {
          key: 'organization--name',
          contents: <Link to={`/organizations/${o.id}`}>{o.name}</Link>,
        },
        {
          key: 'organization--membership',
          contents: 'Owner',
        },
        {
          key: 'organization--actions',
          contents: (
            <ComponentSpacer align={Alignment.Right}>
              <DeleteOrgButton org={o} onDeleteOrg={onDeleteOrg} />
            </ComponentSpacer>
          ),
        },
      ],
    }))
  }

  private get emptyState(): JSX.Element {
    return (
      <EmptyState size={ComponentSize.Large}>
        <EmptyState.Text text="Looks like you are not a member of any Organizations" />
      </EmptyState>
    )
  }
}

export default OrganizationsPageContents
