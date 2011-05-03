// Scrabble move generator. Given a word list, board, and your current tiles,
// outputs all legal moves ranked by point value.

// TODO: Blank tiles! Also, scoring + all-tile-used bonus. And concurrency...

package main

import ("bufio"; "container/vector"; "flag"; "fmt"; "os"; "strings";
        "./moves"; "./sortwith"; "./trie")

var wordListFlag = flag.String(
    "w", "",
    "File with space-separated list of legal words, in upper-case.")
var boardFlag = flag.String(
    "b", "",
    "File with board structure. Format: * indicates starting point, 1 and 2 " +
    "indicate double and triple word score tiles, 3 and 4 indicate double " +
    "and triple letter score tiles, - indicates blank tiles, and upper-case " +
    "letters indicate existing words.")
var tilesFlag = flag.String(
    "t", "", "List of all 7 player tiles, in lower-case.")

func readWordList(wordListFile* os.File) (dict* trie.Trie) {
  wordListReader := bufio.NewReader(wordListFile)
  dict = trie.New()
  for {
    word, err := wordListReader.ReadString(" "[0])
    if err != nil {
      return
    }
    dict.Insert(strings.TrimSpace(word))
  }
  return
}

const boardSize = 15

func readBoard(boardFile* os.File) (board [][]byte) {
  board = make([][]byte, boardSize)
  for i := 0; i < boardSize; i++ {
    board[i] = make([]byte, boardSize)
    _, err := boardFile.Read(board[i])
    if err != nil {
      os.Exit(1)
    }
    _, err = boardFile.Seek(1, 1)
    if err != nil {
      os.Exit(1)
    }
  }
  return
}

func printBoard(board [][]byte) {
  for i := 0; i < boardSize; i++ {
    for j := 0; j < boardSize; j++ {
      fmt.Printf("%c", board[i][j])
    }
    fmt.Printf("\n")
  }
}

func transpose(board [][]byte) (transposedBoard [][]byte) {
  transposedBoard = make([][]byte, boardSize)
  for i := 0; i < boardSize; i++ {
    transposedBoard[i] = make([]byte, boardSize)
    copy(transposedBoard[i], board[i])
  }
  for i := 0; i < boardSize; i++ {
    for j := 0; j < i; j++ {
      transposedBoard[i][j], transposedBoard[j][i] =
          transposedBoard[j][i], transposedBoard[i][j]
    }
  }
  return
}

func existing(char byte) bool {
  return (char >= 'a' && tiles <= 'z') || char == '*'
}

func getTilesInFollowing(dict *trie.Trie, prefix string,
                         tiles []byte) (tilesInFollowing []byte) {
  following := dict.Following(prefix)
  tilesInFollowing = make([]byte, 0, len(following))
  k := 0
  for i := 0; i < len(following); i++ {
    for j := 0; j < len(tiles); j++ {
      if following[i] == tiles[j] {
        tilesInFollowing = tilesInFollowing[:k + 1]
        tilesInFollowing[k] = following[i]
        k++
      }
    }
  }
  return
}

func extendRight(dict* trie.Trie, start *moves.Location, board [][]byte,
                 tiles []byte) (moveList *vector.Vector) {
  prefix = ""
  i := 0
  for ; start.X + i < boardSize && existing(board[start.X][start.Y + i]); i++ {
  }
  prefix = string(board[start.X][start.X : start.X + i])
  following := getTilesInFollowing(dict, prefix, tiles)
  for i := 0; i < len(following); i++ {
    // TODO(ariw): Implement.
  }
  return
}

func extendLeft(dict *trie.Trie, start *moves.Location, board [][]byte,
                tiles []byte) (moveList *vector.Vector) {
  if (start.Y - 1 < 0) {
    return
  }

  // Place a tile in current position (if valid) then extend both directions.
  for i := 0; i < len(tiles); i++ {
    board[start.X][start.Y - 1] = tiles[i]
    if (!isVerticallyValid(dict, board, start.Y - 1)) {
      continue
    }

    // If we've got tiles/space left, attempt to extend left from current
    // location.
    leftStart := copy(start)
    extendLeft(dict, leftStart, board, tiles - tiles[i])
    // TODO(ariw): Implement.

    // Remaining moves can be generated by extending right from current
    // location.
    rightMoveList := extendRight(dict, start, board, tiles)
    moveList.AppendVector(rightMoveList)
  }
  return
}

func getMoveList(dict *trie.Trie, board [][]byte,
                 tiles []byte) (moveList vector.Vector) {
  // Look for lowercase characters as well as * on the board.
  for i := 0; i < boardSize; i++ {
    for j := 0; j < boardSize; j++ {
      if (existing(board[i][j]) {
        moveList.AppendVector(
            extendLeft(dict, &moves.Location{i, j}, board, tiles))
      }
    }
  }
  return
}


func setDirection(direction moves.Direction, moveList *vector.Vector) {
  for i := 0; i < moveList.Len(); i++ {
    move := moveList.At(i).(moves.Move)
    move.Direction = direction
    moveList.Set(i, move)
  }
}

func main() {
  flag.Parse()
  wordListFile, err := os.Open(*wordListFlag, os.O_RDONLY, 0);
  defer wordListFile.Close();
  if err != nil {
    fmt.Printf("need valid file for -w, found %s\n", *wordListFlag)
    os.Exit(1)
  }
  boardFile, err := os.Open(*boardFlag, os.O_RDONLY, 0);
  defer boardFile.Close();
  if err != nil {
    fmt.Printf("need valid file for -b, found %s\n", *boardFlag)
    os.Exit(1)
  }
  if len(*tilesFlag) != 7 {
    fmt.Printf("need 7 tiles in -t, found %d\n", len(*tilesFlag))
    os.Exit(1)
  }
  dict := readWordList(wordListFile)
  board := readBoard(boardFile)
  moveList := getMoveList(dict, board, []byte(*tilesFlag))
  setDirection(moves.RIGHT, &moveList)
  downMoveList := getMoveList(dict, transpose(board), []byte(*tilesFlag))
  setDirection(moves.DOWN, &downMoveList)
  moveList.AppendVector(&downMoveList)
  sortwith.SortWith(moveList, moves.Less)
  for i := 0; i < moveList.Len(); i++ {
    move := moveList.At(i).(moves.Move)
    fmt.Printf("%d. %s, worth %d points, starting at %d, %d, going %d.",
               i, move.Word, move.Score, move.Start.X, move.Start.Y,
               move.Direction)
  }
}

