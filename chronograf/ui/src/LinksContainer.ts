import {Container} from 'unstated'
import {getLinks} from 'src/shared/apis/links'
import {Links} from 'src/types/v2/links'

interface LinkState {
  links: Links
}

export class LinksContainer extends Container<LinkState> {
  public state = {
    links: null,
  }

  public getLinks = async () => {
    const links = await getLinks()
    this.setState({links})
  }
}
