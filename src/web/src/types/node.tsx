export interface NodeAstPart {
  type: string;
  children: NodeAstPart[];
  attributes: { [key: string]: any };
}

export interface Node {
  id: string;
  name: string;
  path: string;
  length: number;
  childNodes: Node[];
  ast: NodeAstPart;
}
