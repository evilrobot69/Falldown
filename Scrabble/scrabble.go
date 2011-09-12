// Scrabble move generator. Given a word list, board, and your current tiles,
// outputs all legal moves ranked by point value.

package main

import ("container/vector"; "flag"; "fmt"; "os";
        "cross_check"; "moves"; "sort_with"; "trie"; "util")

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
    "t", "", "List of all 7 player tiles, in upper-case.")
var letterValuesFlag = flag.String(
    "l", "1 3 3 2 1 4 2 4 1 8 5 1 3 1 1 3 10 1 1 1 1 4 4 8 4 10",
    "Space-separated list of letter point values, from A-Z.")

func transpose(board [][]byte) (transposedBoard [][]byte) {
  transposedBoard = make([][]byte, BOARD_SIZE)
  for i := 0; i < BOARD_SIZE; i++ {
    transposedBoard[i] = make([]byte, BOARD_SIZE)
    copy(transposedBoard[i], board[i])
  }
  for i := 0; i < BOARD_SIZE; i++ {
    for j := 0; j < i; j++ {
      transposedBoard[i][j], transposedBoard[j][i] =
          transposedBoard[j][i], transposedBoard[i][j]
    }
  }
  return
}

// Retrieve the tiles that can possibly follow prefix in dict.
func getTilesInFollowing(dict *trie.Trie, prefix string,
                         tiles map[byte] int) (tilesInFollowing []byte) {
  following := dict.Following(prefix)
  tilesInFollowing = make([]byte, 0, len(following))
  k := 0
  for i := 0; i < len(following); i++ {
    count, _ := tiles[following[i]]
    tilesInFollowing = tilesInFollowing[:k + count]
    for i := 0; i < count; i++ {
      tilesInFollowing[k] = following[i]
      k++
    }
  }
  return
}

func placeTile(dict* trie.Trie, location *moves.Location, board [][]byte,
               tile byte) (newBoard [][]byte, placed bool) {
  // Ensure that location for tile is vertically valid.
  /*i := location.Y - 1
  for ; util.Existing(board, &moves.Location({location.X, i})); i--
  i++*/
  return nil, true
}

func extendRight(dict* trie.Trie, start moves.Location, board [][]byte,
                 tiles map[byte] int) (moveList *vector.Vector) {
  if len(tiles) == 0 || start.Y == BOARD_SIZE {
    return
  }
  // Place a tile (if extensions exist to prefix and is valid), then recurse.
  i := 0
  for ; util.Existing(board, &moves.Location{start.X, start.Y + i}); i++ {
  }
  following := getTilesInFollowing(
      dict, string(board[start.X][start.X : start.X + i]), tiles)
  moveList = new(vector.Vector)
  for j := 0; j < len(following); j++ {
    // TODO: Also check to see if after placeTile completes we have a valid
    // move. Then save it to moveList.
    newBoard, placed := placeTile(dict, &moves.Location{start.X, start.Y + 1},
                                  board, following[j])
    if !placed {
      continue
    }
    newTiles := copy(tiles)
    if following[j] == 1:
      newTiles[j] = 0, false
    else:
      newTiles[j] = following[j] - 1, true
    extendRight(dict, start, newBoard, newTiles)
  }
  return
}

func extendLeft(dict *trie.Trie, start moves.Location, board [][]byte,
                tiles map[byte] int) (moveList *vector.Vector) {
  if len(tiles) == 0 || start.Y < 0 {
    return
  }

  moveList = new(vector.Vector)

  // If there are possible right extensions, extend right.
  rightMoveList := extendRight(dict, start, board, tiles)
  moveList.AppendVector(rightMoveList)


  // If it is valid, place a tile at current location and extend left.
  for tile, _ := range tiles {
    // TODO: Also check to see if after placeTile completes we have a valid
    // move. Then save it to moveList.
    newBoard, placed := placeTile(dict, &moves.Location{start.X, start.Y - 1},
                                  board, tile)
    if !placed {
      continue
    }
    newTiles := copy(tiles)
    newTiles[tile] = false, false
    // Extend left from current location.
    extendLeft(dict, moves.Location{start.X, start.Y - 1}, newBoard, newTiles)
  }
  return
}

func getMoveList(dict *trie.Trie, board [][]byte, tiles map[byte] int,
                 letterValues map[byte] int) (moveList vector.Vector) {
  // Look for lowercase characters as well as * on the board.
  for i := 0; i < BOARD_SIZE; i++ {
    for j := 0; j < BOARD_SIZE; j++ {
      if util.Existing(board, &moves.Location{i, j}) {
        moveList.AppendVector(
            extendLeft(dict, moves.Location{i - 1, j}, board, tiles))
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
  dict := util.ReadWordList(wordListFile)
  board := util.ReadBoard(boardFile)
  tiles := util.ReadTiles(*tilesFlag)
  letterValues := util.ReadLetterValues(*letterValuesFlag)

  // Get moves going both right and down.
  crossCheckSet := cross_check.GetCrossChecks(dict, board, tiles, letterValues)
  moveList := getMoveList(dict, board, tiles, letterValues)
  setDirection(moves.RIGHT, &moveList)
  transposedBoard := transpose(board)
  downCrossCheckSet := cross_check.GetCrossChecks(dict, transposedBoard, tiles,
                                                  letterValues)
  downMoveList := getMoveList(dict, transpose(board), tiles, letterValues)
  setDirection(moves.DOWN, &downMoveList)
  moveList.AppendVector(&downMoveList)
  sort_with.SortWith(moveList, moves.Less)
  for i := 0; i < moveList.Len(); i++ {
    move := moveList.At(i).(moves.Move)
    fmt.Printf("%d. %s, worth %d points, starting at %d, %d, going %d.",
               i, move.Word, move.Score, move.Start.X, move.Start.Y,
               move.Direction)
  }
}

